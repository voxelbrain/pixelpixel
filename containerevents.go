package main

import (
	"archive/tar"
)

type ContainerEvents struct {
	c chan *Event
	ContainerManager
}

type Event struct {
	Type EventType
	Id   ContainerId
}

type EventType int

const (
	EventContainerCreated = EventType(iota)
	EventContainerDestroyed
)

func NewContainerEvents(cm ContainerManager) (ContainerManager, <-chan *Event) {
	c := make(chan *Event)
	return &ContainerEvents{
		c:                c,
		ContainerManager: cm,
	}, c
}

func (ce *ContainerEvents) NewContainer(fs *tar.Reader, envInjection []string) (ContainerId, error) {
	id, err := ce.ContainerManager.NewContainer(fs, envInjection)
	if err == nil {
		ce.c <- &Event{
			Type: EventContainerCreated,
			Id:   id,
		}
	}
	return id, err
}

func (ce *ContainerEvents) DestroyContainer(id ContainerId) error {
	err := ce.ContainerManager.DestroyContainer(id)
	if err == nil {
		ce.c <- &Event{
			Type: EventContainerDestroyed,
			Id:   id,
		}
	}
	return err
}
