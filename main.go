package main

import (
	"log"
	"sync"
)

type PubSub[T any] struct {
	subscribers []chan T
	mu          sync.RWMutex
	closed      bool
}

func NewPubSub[T any]() *PubSub[T] {
	return &PubSub[T]{
		mu: sync.RWMutex{},
	}
}

func (s *PubSub[T]) Subscribe() <-chan T {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	r := make(chan T)

	s.subscribers = append(s.subscribers, r)

	return r
}

func (s *PubSub[T]) Publish(value T) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}

	for _, ch := range s.subscribers {
		ch <- value
	}
}

func (s *PubSub[T]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	for _, ch := range s.subscribers {
		close(ch)
	}

	s.closed = true
}

func main() {
	ps := NewPubSub[string]()

	wg := sync.WaitGroup{}

	s1 := ps.Subscribe()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for msg := range s1 {
			log.Println("subscriber 1, value ", msg)
		}
		log.Println("subscriber 1, exiting")
		
	}()

	s2 := ps.Subscribe()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for val := range s2 {
			log.Println("subscriber 2, value ", val)
		}

		log.Println("subscriber 2, exiting")
	}()

	ps.Publish("one")
	ps.Publish("two")
	ps.Publish("three")

	ps.Close()

	wg.Wait()

	log.Println("completed")
}