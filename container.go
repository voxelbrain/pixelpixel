package main

import (
	"archive/tar"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type ContainerId string

func (c ContainerId) String() string {
	return string(c)
}

type ContainerManager interface {
	NewContainer(fs *tar.Reader, envInjection []string) (ContainerId, error)
	DestroyContainer(id ContainerId) error
	WaitFor(id ContainerId) chan bool
	Logs(id ContainerId) ([]byte, error)
	SocketAddress(id ContainerId) (string, error)
}

type ContainerManagerAPI struct {
	*sync.Mutex
	idMap map[string]ContainerId
	http.Handler
	ContainerManager
}

func NewContainerManagerAPI(cm ContainerManager) *ContainerManagerAPI {
	handler := mux.NewRouter()
	cma := &ContainerManagerAPI{
		Mutex:            &sync.Mutex{},
		idMap:            map[string]ContainerId{},
		Handler:          handler,
		ContainerManager: cm,
	}

	handler.Path("/").Methods("POST").HandlerFunc(cma.CreatePixelHandler)
	handler.Path("/{key}").Methods("PUT").HandlerFunc(cma.UpdatePixelHandler)
	handler.Path("/{key}").Methods("DELETE").HandlerFunc(cma.DeletePixelHandler)
	handler.Path("/{key}").Methods("GET").HandlerFunc(cma.GetPixelLogHandler)

	return cma
}

func (cma *ContainerManagerAPI) CreatePixelHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fs := tar.NewReader(r.Body)

	id, err := cma.ContainerManager.NewContainer(fs, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := cma.generateId()
	cma.Lock()
	defer cma.Unlock()
	cma.idMap[key] = id

	fmt.Fprintf(w, "%s", key)
}

func (cma *ContainerManagerAPI) UpdatePixelHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		http.Error(w, "Key missing", http.StatusBadRequest)
		return
	}

	id, ok := cma.idMap[key]
	if !ok {
		http.Error(w, "Unknown key", http.StatusBadRequest)
		return
	}

	err := cma.DestroyContainer(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fs := tar.NewReader(r.Body)
	id, err = cma.ContainerManager.NewContainer(fs, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cma.Lock()
	defer cma.Unlock()
	cma.idMap[key] = id

	http.Error(w, "", http.StatusNoContent)
}

const (
	chars  = `abcdefghijklmnopqrstuvwxyz1234567890`
	length = 3
)

func (cma *ContainerManagerAPI) generateId() string {
	key := make([]byte, length)
	idx := rand.Perm(len(chars))
	for i := 0; i < length; i++ {
		key[i] = chars[idx[i]]
	}
	return string(key)
}

func (cma *ContainerManagerAPI) GetPixelLogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		http.Error(w, "Key missing", http.StatusBadRequest)
		return
	}

	id, ok := cma.idMap[key]
	if !ok {
		http.Error(w, "Unknown key", http.StatusBadRequest)
		return
	}

	logs, err := cma.Logs(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write(logs)
}

func (cma *ContainerManagerAPI) DeletePixelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		http.Error(w, "Key missing", http.StatusBadRequest)
	}

	_, ok = cma.idMap[key]
	if !ok {
		http.Error(w, "Unknown key", http.StatusBadRequest)
		return
	}

	cma.Lock()
	defer cma.Unlock()
	delete(cma.idMap, key)
}
