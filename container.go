package main

import (
	"archive/tar"
	"net/http"
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
}

type ContainerManagerAPI struct {
}

func NewContainerManagerAPI(cm ContainerManager) *ContainerManagerAPI {
	cma := &ContainerManagerAPI{}
	return cma
}

func (cma *ContainerManagerAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
