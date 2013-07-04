package main

import (
	"archive/tar"
	"fmt"
	"net/http"

	"github.com/surma-dump/gouuid"

	"github.com/gorilla/mux"
)

type ContainerId string

func (c ContainerId) String() string {
	return string(c)
}

type ContainerManager interface {
	NewContainer(fs *tar.Reader, envInjection []string) (ContainerId, error)
	DestroyContainer(id ContainerId) error
	WaitFor(id ContainerId) chan bool
	Logs(id ContainerId) ([]byte, error)
	Port(id ContainerId) (int, error)
}

type ContainerManagerAPI struct {
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
	handler.Path("/{id}").Methods("DELETE").HandlerFunc(cma.DeletePixelHandler)
	handler.Path("/{id}").Methods("GET").HandlerFunc(cma.GetPixelLogHandler)
	handler.Path("/{id}/").HandlerFunc(cma.ReverseProxy)

	return cma
}

func (cma *ContainerManagerAPI) CreatePixelHandler(w http.ResponseWriter, r *http.Request) {
	fs := tar.NewReader(r.Body)
	defer r.Body.Close()

	id, err := cma.ContainerManager.NewContainer(fs, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "%s", id)
}

func (cma *ContainerManagerAPI) GetPixelLogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := gouuid.ParseString(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	logs, err := cma.Logs(ContainerId(id.String()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write(logs)
}

func (cma *ContainerManagerAPI) DeletePixelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := gouuid.ParseString(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	cma.ContainerManager.DestroyContainer(ContainerId(id.String()))
	http.Error(w, "", http.StatusNoContent)
}

func (cma *ContainerManagerAPI) ReverseProxy(w http.ResponseWriter, r *http.Request) {}
