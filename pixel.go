package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

type PixelApi struct {
	*sync.RWMutex
	container map[string]ContainerId
	pixels    map[string]*bytes.Buffer
	cm        ContainerManager
	Messages  chan *Message
	http.Handler
}

func NewPixelApi(cm ContainerManager) *PixelApi {
	pa := &PixelApi{
		RWMutex:   &sync.RWMutex{},
		Messages:  make(chan *Message),
		container: make(map[string]ContainerId),
		pixels:    make(map[string]*bytes.Buffer),
		cm:        cm,
	}
	h := mux.NewRouter()
	h.PathPrefix("/").Methods("POST").HandlerFunc(pa.CreatePixel)
	h.PathPrefix("/{id}/").Methods("PUT").HandlerFunc(pa.ValidatePixelId(pa.UpdatePixel))
	// h.PathPrefix("/{id}/").Methods("DELETE").Handler(pa.CreatePixel)
	h.PathPrefix("/{id}/content").Methods("GET").HandlerFunc(pa.ValidatePixelId(pa.GetPixelContent))
	h.PathPrefix("/{id}/logs").Methods("GET").HandlerFunc(pa.ValidatePixelId(pa.GetPixelLogs))
	pa.Handler = h
	return pa
}

func (pa *PixelApi) CreatePixel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := pa.generateId()

	buf := &bytes.Buffer{}
	io.Copy(buf, r.Body)
	if buf.Len() <= 0 {
		http.Error(w, "Empty fs", http.StatusBadRequest)
		return
	}
	cid, err := pa.cm.NewContainer(tar.NewReader(bytes.NewReader(buf.Bytes())), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pa.Lock()
	pa.container[id] = cid
	pa.pixels[id] = &bytes.Buffer{}
	pa.Unlock()

	pa.Messages <- &Message{
		Pixel: id,
		Type:  TypeCreate,
	}

	go pa.pixelListener(id)

	http.Error(w, id, http.StatusCreated)
}

func (pa *PixelApi) UpdatePixel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := mux.Vars(r)["id"]

	pa.RLock()
	cid := pa.container[id]
	pa.RUnlock()

	buf := &bytes.Buffer{}
	io.Copy(buf, r.Body)
	if buf.Len() <= 0 {
		http.Error(w, "Empty fs", http.StatusBadRequest)
		return
	}

	pa.cm.DestroyContainer(cid, true)
	<-pa.cm.WaitFor(cid)

	cid, err := pa.cm.NewContainer(tar.NewReader(bytes.NewReader(buf.Bytes())), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pa.Lock()
	pa.container[id] = cid
	pa.Unlock()

	pa.Messages <- &Message{
		Pixel: id,
		Type:  TypeChange,
	}

	go pa.pixelListener(id)

	http.Error(w, id, http.StatusCreated)
}

func (pa *PixelApi) pixelListener(id string) {
	pa.RLock()
	cid := pa.container[id]
	buf := pa.pixels[id]
	pa.RUnlock()

	time.Sleep(1 * time.Second)
	addr, err := pa.cm.SocketAddress(cid)
	if err != nil {
		pa.Messages <- &Message{
			Pixel:   id,
			Type:    TypeFailure,
			Payload: fmt.Sprintf("Could not get socket address of %s: %s", id, err),
		}
		return
	}
	c, err := net.Dial("tcp", addr)
	if err != nil {
		pa.Messages <- &Message{
			Pixel:   id,
			Type:    TypeFailure,
			Payload: fmt.Sprintf("Could not connect to pixel %s: %s", id, err),
		}
		return
	}
	defer c.Close()

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
		buf.Reset()
		io.Copy(buf, tr)
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
		_, ok = pa.container[id]
		pa.RUnlock()
		if !ok {
			http.Error(w, "No such pixel", http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (pa *PixelApi) GetPixelLogs(w http.ResponseWriter, r *http.Request) {
	cid := pa.container[mux.Vars(r)["id"]]

	logs, err := pa.cm.Logs(cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(logs)
}

func (pa *PixelApi) GetPixelContent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	pa.RLock()
	buf := pa.pixels[id]
	pa.RUnlock()

	w.Header().Set("cache-control", "private, max-age=0, no-cache")
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	io.Copy(w, bytes.NewReader(buf.Bytes()))
}

const (
	chars  = `abcdefghijklmnopqrstuvwxyz1234567890`
	length = 3
)

func (pa *PixelApi) generateId() string {
	key := make([]byte, length)
	idx := rand.Perm(len(chars))
	for i := 0; i < length; i++ {
		key[i] = chars[idx[i]]
	}
	return string(key)
}
