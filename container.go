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
	AllContainers() []ContainerId
	DestroyContainer(id ContainerId) error
	WaitFor(id ContainerId) chan bool
	Logs(id ContainerId) ([]byte, error)
	SocketAddress(id ContainerId) (string, error)
}

type ContainerManagerAPI struct {
	*sync.Mutex
	http.Handler
	ContainerManager
}

func NewContainerManagerAPI(cm ContainerManager) *ContainerManagerAPI {
	handler := mux.NewRouter()
	cma := &ContainerManagerAPI{
		Handler:          handler,
		ContainerManager: cm,
	}

	handler.Path("/").Methods("POST").HandlerFunc(cma.CreatePixelHandler)
	handler.Path("/{id}/").Methods("DELETE").HandlerFunc(cma.DeletePixelHandler)
	handler.Path("/{id}/").Methods("GET").HandlerFunc(cma.GetPixelLogHandler)

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

	fmt.Fprintf(w, "%s", id)
}

func (cma *ContainerManagerAPI) GetPixelLogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.Error(w, "id missing", http.StatusBadRequest)
		return
	}

	logs, err := cma.Logs(ContainerId(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write(logs)
}

func (cma *ContainerManagerAPI) DeletePixelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.Error(w, "id missing", http.StatusBadRequest)
	}

	err := cma.DestroyContainer(ContainerId(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, "", http.StatusNoContent)
}
