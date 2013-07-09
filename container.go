package main

import (
	"archive/tar"
	"math/rand"
	"time"
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
	DestroyContainer(id ContainerId, purge bool) error
	WaitFor(id ContainerId) chan bool
	Logs(id ContainerId) ([]byte, error)
	SocketAddress(id ContainerId) (string, error)
}
