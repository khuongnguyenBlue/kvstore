package storage

import (
	"fmt"
	"sync"
	"time"
)

type MemoryStore struct {
	data map[string]string
	ttl  map[string]int64
	mu   sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]string),
		ttl:  make(map[string]int64),
	}
}

func (m *MemoryStore) Get(key string) (string, bool) {
	m.mu.RLock()

	if m.isExpired(key) {
		m.mu.RUnlock()

		m.mu.Lock()
		defer m.mu.Unlock()

		if m.isExpired(key) {
			m.delete(key)
		}

		return "", false
	}

	val, found := m.data[key]
	m.mu.RUnlock()
	return val, found
}

func (m *MemoryStore) Set(key, value string, ttlSeconds *int64) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value

	if ttlSeconds != nil && *ttlSeconds > 0 {
		m.ttl[key] = time.Now().Unix() + *ttlSeconds
	} else {
		delete(m.ttl, key)
	}

	return nil
}

func (m *MemoryStore) Delete(key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	_, existed := m.data[key]
	m.delete(key)

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

		if m.isExpired((k)) {
			continue
		}

		result[k] = v
	}

	return result, nil
}

func (m *MemoryStore) isExpired(key string) bool {
	expiration, hasExpiration := m.ttl[key]
	if !hasExpiration {
		return false
	}

	return time.Now().Unix() >= expiration
}

func (m *MemoryStore) delete(key string) {
	delete(m.data, key)
	delete(m.ttl, key)
}
