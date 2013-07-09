package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/voxelbrain/pixelpixel/protocol"
)

type PixelApi struct {
	*sync.RWMutex
	container map[string]ContainerId
	pixels    map[string]*bytes.Buffer
	cm        ContainerManager
	Messages  chan *protocol.Message
	http.Handler
}

func NewPixelApi(cm ContainerManager) *PixelApi {
	pa := &PixelApi{
		RWMutex:   &sync.RWMutex{},
		Messages:  make(chan *protocol.Message),
		container: make(map[string]ContainerId),
		pixels:    make(map[string]*bytes.Buffer),
		cm:        cm,
	}
	h := mux.NewRouter()
	h.PathPrefix("/").Methods("POST").HandlerFunc(pa.CreatePixel)
	// h.PathPrefix("/{id}/").Methods("PUT").Handler(pa.CreatePixel)
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
	pixelbuf := &bytes.Buffer{}
	pa.pixels[id] = pixelbuf
	pa.Unlock()

	go func() {
		time.Sleep(1 * time.Second)
		addr, err := pa.cm.SocketAddress(cid)
		if err != nil {
			pa.Messages <- &protocol.Message{
				Pixel:   id,
				Type:    protocol.TypeFailure,
				Payload: fmt.Sprintf("Could not get socket address of %s: %s", id, err),
			}
			return
		}
		c, err := net.Dial("tcp", addr)
		if err != nil {
			pa.Messages <- &protocol.Message{
				Pixel:   id,
				Type:    protocol.TypeFailure,
				Payload: fmt.Sprintf("Could not connect to pixel %s: %s", id, err),
			}
			return
		}
		defer c.Close()

		pa.Messages <- &protocol.Message{
			Pixel: id,
			Type:  protocol.TypeCreated,
		}

		tr := tar.NewReader(c)
		for {
			_, err := tr.Next()
			if err != nil {
				break
			}
			pixelbuf.Reset()
			io.Copy(pixelbuf, tr)
			pa.Messages <- &protocol.Message{
				Pixel: id,
				Type:  protocol.TypeChange,
			}
		}
	}()

	http.Error(w, id, http.StatusCreated)
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
