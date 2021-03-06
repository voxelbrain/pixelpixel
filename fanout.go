package main

import (
	"reflect"
	"sync"
)

// Fanout is a small structure implementing a more or less generic,
// thread-safe fanout. Fanouts are created on input channels and
// propagate each received value to all consumers in order.
// Consumers can close their channels.
type Fanout struct {
	*sync.RWMutex
	consumers map[chan interface{}]struct{}
	closing   map[chan interface{}]bool
}

// Create a new fanout from a channel. c has to be a channel type
// with receiving capabilities.
func NewFanout(c interface{}) *Fanout {
	f := &Fanout{
		RWMutex:   &sync.RWMutex{},
		consumers: map[chan interface{}]struct{}{},
		closing:   map[chan interface{}]bool{},
	}
	go f.loop(c)
	return f
}

// Create a new consumer output.
func (f *Fanout) Output() <-chan interface{} {
	c := make(chan interface{})
	f.Lock()
	defer f.Unlock()
	f.consumers[c] = struct{}{}
	return c
}

// Close a consumer channel, stopping propagation for this particular
// consumer.
func (f *Fanout) Close(rc <-chan interface{}) {
	f.RLock()
	defer f.RUnlock()

	// Lookup original channel because we can't call close()
	// on a receive-only channel
	var c chan interface{}
	for i := range f.consumers {
		if i == rc {
			c = i
		}
	}

	// If channel is not in consumers map are is already about to close
	// don't try to do it again.
	if _, ok := f.closing[c]; c == nil || ok {
		return
	}
	f.closing[c] = true

	// Wait for the current broadcast to finish (effectively unlocking
	// the mutex) and delete the consumer from the map.
	go func() {
		f.Lock()
		defer f.Unlock()
		delete(f.consumers, c)
		delete(f.closing, c)
		close(c)
	}()

	// Eat the values possibly left in channel in case the consumer
	// doesn't.
	go func() {
		for {
			_, ok := <-c
			// If the channel is closed it has been removed from
			// the consumers map by the previous goroutine. Stop eating.
			if !ok {
				return
			}
		}
	}()
}

func (f *Fanout) loop(c interface{}) {
	ch := reflect.ValueOf(c)
	if ch.Type().Kind() != reflect.Chan {
		panic("Not a channel type")
	}
	if ch.Type().ChanDir()&reflect.RecvDir == 0 {
		panic("Cannot receive on given channel")
	}
	for {
		v, ok := ch.Recv()
		if !ok {
			f.closeConsumers()
			return
		}
		f.broadcast(v.Interface())
	}
}

func (f *Fanout) closeConsumers() {
	f.RLock()
	defer f.RUnlock()
	for c := range f.consumers {
		f.Close(c)
	}
}

func (f *Fanout) broadcast(v interface{}) {
	f.RLock()
	defer f.RUnlock()
	for c := range f.consumers {
		c <- v
	}
}
