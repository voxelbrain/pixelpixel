package main

import (
	"archive/tar"
	"math/rand"
	"net"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Container interface {
	// Returns true if the process inside the container is running
	IsRunning() bool
	// Returns the address to dial to connect to the container
	Address() net.Addr
	// Buffer of the cumulated output of the container
	Logs() string
	// Sends a soft kill signal (a la SIGINT)
	SoftKill()
	// Sends a hard kill signal (a la SIGKILL)
	HardKill()
	// Waits for the process inside the container to finsh
	Wait()
	// Clean the containers directory. Cleanup waits for the container
	// to terminate before starting the cleanup.
	Cleanup()
}

type ContainerCreator interface {
	CreateContainer(fs *tar.Reader, envInjection []string) (Container, error)
}
