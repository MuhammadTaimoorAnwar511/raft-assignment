package kvstore

import (
	"sync"

	"github.com/MuhammadTaimoorAnwar511/raft-assignment/internal"
)

// Store represents an in-memory key-value store.
type Store struct {
	mu    sync.RWMutex
	kvMap map[string]string
}

// NewStore creates and returns a new Store.
func NewStore() *Store {
	return &Store{
		kvMap: make(map[string]string),
	}
}

// Put sets the value for a given key.
func (s *Store) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.kvMap[key] = value
}

// Append appends the value to an existing key (or behaves like Put if the key doesn't exist).
func (s *Store) Append(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	oldVal, exists := s.kvMap[key]
	if !exists {
		s.kvMap[key] = value
	} else {
		s.kvMap[key] = oldVal + value
	}
}

// Get retrieves the value for a given key. The boolean indicates if the key was found.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.kvMap[key]
	return val, ok
}

// ApplyCommand applies a Raft command (Put, Append, or Get).
// In practice, "Get" doesn't change state, so we might not do anything for that case.
func (s *Store) ApplyCommand(cmd internal.Command) {
	switch cmd.Type {
	case internal.CommandPut:
		s.Put(cmd.Key, cmd.Value)
	case internal.CommandAppend:
		s.Append(cmd.Key, cmd.Value)
	case internal.CommandGet:
		// Get is read-only, so no state change. We do nothing here.
		// Typically, Get wouldn't be replicated in Raft unless it matters for linearizability.
	}
}
