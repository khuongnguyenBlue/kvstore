package storage

import (
	"fmt"
	"sync"
)

type MemoryStore struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]string),
	}
}

func (m *MemoryStore) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, found := m.data[key]
	return val, found
}

func (m *MemoryStore) Set(key, value string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value

	return nil
}

func (m *MemoryStore) Delete(key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	_, existed := m.data[key]
	delete(m.data, key)

	return existed, nil
}

func (m *MemoryStore) List(limit int) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]string)

	for k, v := range m.data {
		if len(result) >= limit {
			break
		}

		result[k] = v
	}

	return result, nil
}
