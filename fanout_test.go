package main

import (
	"sync"
	"testing"
	"time"
)

const (
	numConsumers = 3
	numSends     = 3
)

func TestBroadcast(t *testing.T) {
	c := make(chan int)

	f := NewFanout(c)
	wg := &sync.WaitGroup{}
	corrects := make([]bool, numConsumers)
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			corrects[idx] = true
			c := f.Output()

			is := make([]int, 0, numSends)
			for i := 0; i < numSends; i++ {
				is = append(is, (<-c).(int))
			}
			for i, v := range is {
				if i != v {
					t.Fatalf("Consumer %d received invalid value %d on position %d", idx, v, i)
				}
			}
		}(i)
	}

	go func() {
		for i := 0; i < numSends; i++ {
			c <- i
		}
	}()

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	select {
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout")
	case <-done:
	}
}

func TestClosingConsumer(t *testing.T) {
	c := make(chan int)
	f := NewFanout(c)
	wg := &sync.WaitGroup{}
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c := f.Output()
			for i := 0; i < numSends; i++ {
				<-c
			}
		}(i)
	}

	// Bastard consumer, closing early
	wg.Add(1)
	go func() {
		defer wg.Done()
		c := f.Output()
		f.Close(c)
	}()

	go func() {
		for i := 0; i < numSends; i++ {
			c <- i
		}
	}()

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	select {
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout")
	case <-done:
	}
}

func TestClosingProducer(t *testing.T) {
	c := make(chan int)
	f := NewFanout(c)
	wg := &sync.WaitGroup{}
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c := f.Output()
			for _ = range c {
			}
		}(i)
	}

	go func() {
		for i := 0; i < numSends; i++ {
			c <- i
		}
		close(c)
	}()

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	select {
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout")
	case <-done:
	}
}
