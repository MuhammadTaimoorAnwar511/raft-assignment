package kvstore

import "sync"

type Store struct {
    mu    sync.Mutex
    data  map[string]string
}

func NewStore() *Store {
    return &Store{
        data: make(map[string]string),
    }
}

// Put overwrites the existing value for the key
func (s *Store) Put(key, value string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
}

// Append concatenates value to the existing value for the key
func (s *Store) Append(key, value string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] += value
}

// Get retrieves the value for the key
func (s *Store) Get(key string) (string, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    val, ok := s.data[key]
    return val, ok
}
