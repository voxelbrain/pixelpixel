package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Pixel struct {
	Id string
	Container
	Filesystem map[string]string
	LastImage  *bytes.Buffer
}

type PixelApi struct {
	*sync.RWMutex
	pixels   map[string]*Pixel
	cc       ContainerCreator
	Messages chan *Message
	http.Handler
}

func NewPixelApi(cc ContainerCreator) *PixelApi {
	pa := &PixelApi{
		RWMutex:  &sync.RWMutex{},
		Messages: make(chan *Message),
		pixels:   make(map[string]*Pixel),
		cc:       cc,
	}
	h := mux.NewRouter()
	h.PathPrefix("/").Methods("POST").HandlerFunc(pa.CreatePixel)
	h.PathPrefix("/{id}/content").Methods("GET").HandlerFunc(pa.ValidatePixelId(pa.GetPixelContent))
	h.PathPrefix("/{id}/logs").Methods("GET").HandlerFunc(pa.ValidatePixelId(pa.GetPixelLogs))
	h.PathPrefix("/{id}/fs").Methods("GET").HandlerFunc(pa.ValidatePixelId(pa.GetPixelFs))
	h.PathPrefix("/{id}/").Methods("GET").HandlerFunc(pa.ValidatePixelId(pa.CheckPixel))
	h.PathPrefix("/{id}/").Methods("PUT").HandlerFunc(pa.ValidatePixelId(pa.UpdatePixel))
	pa.Handler = h
	return pa
}

func (pa *PixelApi) CreatePixel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := GenerateAlnumString(3)

	buf := &bytes.Buffer{}
	io.Copy(buf, r.Body)
	if buf.Len() <= 0 {
		http.Error(w, "Empty fs", http.StatusBadRequest)
		return
	}
	ctr, err := pa.cc.CreateContainer(tar.NewReader(bytes.NewReader(buf.Bytes())), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pixel := &Pixel{
		Id:         id,
		Container:  ctr,
		Filesystem: fsObject(tar.NewReader(bytes.NewReader(buf.Bytes()))),
		LastImage:  blackPixel,
	}

	pa.Lock()
	pa.pixels[id] = pixel
	pa.Unlock()

	pa.Messages <- &Message{
		Pixel: id,
		Type:  TypeCreate,
	}

	go pa.pixelListener(pixel)

	http.Error(w, id, http.StatusCreated)
}

func (pa *PixelApi) UpdatePixel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := mux.Vars(r)["id"]

	pa.RLock()
	pixel := pa.pixels[id]
	pa.RUnlock()

	buf := &bytes.Buffer{}
	io.Copy(buf, r.Body)
	if buf.Len() <= 0 {
		http.Error(w, "Empty fs", http.StatusBadRequest)
		return
	}

	ctr, err := pa.cc.CreateContainer(tar.NewReader(bytes.NewReader(buf.Bytes())), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	StopContainer(pixel.Container)
	pixel.Container = ctr
	pixel.Filesystem = fsObject(tar.NewReader(bytes.NewReader(buf.Bytes())))
	pixel.LastImage = blackPixel

	pa.Messages <- &Message{
		Pixel: id,
		Type:  TypeChange,
	}

	go pa.pixelListener(pixel)

	http.Error(w, id, http.StatusCreated)
}

func (pa *PixelApi) pixelListener(pixel *Pixel) {
	id := pixel.Id
	// FIXME: Arbitrary wait to accomodate startup
	time.Sleep(1 * time.Second)
	addr := pixel.Address()
	c, err := net.Dial("tcp", addr.String())
	if err != nil {
		pa.Messages <- &Message{
			Pixel:   id,
			Type:    TypeFailure,
			Payload: fmt.Sprintf("Could not connect to pixel %s: %s", id, err),
		}
		return
	}
	defer c.Close()

	go func() {
		pixel.Wait()
		pa.Messages <- &Message{
			Pixel:   id,
			Type:    TypeFailure,
			Payload: fmt.Sprintf("Pixel %s terminated", id),
		}
	}()

	tr := tar.NewReader(c)
	for {
		_, err := tr.Next()
		if err != nil {
			log.Printf("Pixel %s closed its reader: %s", id, err)
			pa.Messages <- &Message{
				Pixel:   id,
				Type:    TypeFailure,
				Payload: err.Error(),
			}
			return
		}
		pixel.LastImage.Reset()
		io.Copy(pixel.LastImage, tr)
		pa.Messages <- &Message{
			Pixel: id,
			Type:  TypeChange,
		}
	}
}

func (pa *PixelApi) ValidatePixelId(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			http.Error(w, "Pixel ID missing", http.StatusBadRequest)
			return
		}

		pa.RLock()
		_, ok = pa.pixels[id]
		pa.RUnlock()
		if !ok {
			http.Error(w, "No such pixel", http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (pa *PixelApi) CheckPixel(w http.ResponseWriter, r *http.Request) {
	pa.RLock()
	pixel := pa.pixels[mux.Vars(r)["id"]]
	pa.RUnlock()

	if pixel.IsRunning() {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (pa *PixelApi) GetPixelLogs(w http.ResponseWriter, r *http.Request) {
	pa.RLock()
	pixel := pa.pixels[mux.Vars(r)["id"]]
	pa.RUnlock()

	io.WriteString(w, pixel.Logs())
}

func (pa *PixelApi) GetPixelFs(w http.ResponseWriter, r *http.Request) {
	pa.RLock()
	pixel := pa.pixels[mux.Vars(r)["id"]]
	pa.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pixel.Filesystem)
}

func (pa *PixelApi) GetPixelContent(w http.ResponseWriter, r *http.Request) {
	pa.RLock()
	pixel := pa.pixels[mux.Vars(r)["id"]]
	pa.RUnlock()

	w.Header().Set("cache-control", "private, max-age=0, no-cache")
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pixel.LastImage.Len()))
	io.Copy(w, bytes.NewReader(pixel.LastImage.Bytes()))
}

func fsObject(r *tar.Reader) map[string]string {
	fs := map[string]string{}
	for {
		hdr, err := r.Next()
		if err != nil {
			break
		}
		if !strings.HasSuffix(hdr.Name, ".go") {
			continue
		}
		buf := &bytes.Buffer{}
		io.Copy(buf, r)
		fs[hdr.Name] = buf.String()
	}
	return fs
}
