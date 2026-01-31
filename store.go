package valmux

import (
	"sync"
)

// Store contains key-value where key is a subject and value is a ValMux.
type Store[T comparable] struct {
	mux     sync.RWMutex
	current map[T]*ValMux

	opts []Option
}

// NewStore creates new Store with ValMux for each key.
func NewStore[T comparable](opts ...Option) *Store[T] {
	return &Store[T]{
		current: make(map[T]*ValMux),
		opts:    opts,
	}
}

// Get return ValMux for the key.
func (s *Store[T]) Get(key T) *ValMux {
	s.mux.RLock()
	vm, ok := s.current[key]
	s.mux.RUnlock()

	if !ok {
		vm = New(DefaultMax, s.opts...)
		s.mux.Lock()
		s.current[key] = vm
		s.mux.Unlock()
	}

	return vm
}

// Set stores ValMux for the key.
func (s *Store[T]) Set(key T, vm *ValMux) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.current[key] = vm
}

// GetAll return copy of the internal storage.
func (s *Store[T]) GetAll() map[T]*ValMux {
	s.mux.RLock()
	defer s.mux.RUnlock()

	m := make(map[T]*ValMux, len(s.current))
	for k, v := range s.current {
		m[k] = v
	}

	return m
}

// Drop removes key from the storage.
func (s *Store[T]) Drop(key T) {
	s.mux.Lock()
	defer s.mux.Unlock()

	delete(s.current, key)
}

// Clear clears the storage.
func (s *Store[T]) Clear() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.current = make(map[T]*ValMux)
}
