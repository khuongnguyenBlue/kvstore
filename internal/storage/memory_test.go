package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test basic Set operation
func TestMemoryStore_Set(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		ttl         *int64
		expectError bool
	}{
		{
			name:        "valid key value",
			key:         "key1",
			value:       "value1",
			ttl:         nil,
			expectError: false,
		},
		{
			name:        "empty key",
			key:         "",
			value:       "value1",
			ttl:         nil,
			expectError: true,
		},
		{
			name:        "empty value",
			key:         "key1",
			value:       "",
			ttl:         nil,
			expectError: false,
		},
		{
			name:        "with TTL",
			key:         "key1",
			value:       "value1",
			ttl:         int64Ptr(60),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()

			err := store.Set(tt.key, tt.value, tt.ttl)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test basic Get operation
func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore()

	// Test non-existent key
	value, found := store.Get("nonexistent")
	assert.False(t, found)
	assert.Empty(t, value)

	// Test existing key
	err := store.Set("key1", "value1", nil)
	require.NoError(t, err)

	value, found = store.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)
}

// Test Set then Get workflow
func TestMemoryStore_SetGet(t *testing.T) {
	store := NewMemoryStore()

	testCases := map[string]string{
		"simple":      "value1",
		"with spaces": "value with spaces",
		"empty":       "",
		"unicode":     "こんにちは",
		"numbers":     "12345",
	}

	// Set all values
	for key, value := range testCases {
		err := store.Set(key, value, nil)
		require.NoError(t, err)
	}

	// Get all values
	for key, expectedValue := range testCases {
		value, found := store.Get(key)
		assert.True(t, found, "Key %s should exist", key)
		assert.Equal(t, expectedValue, value, "Value for key %s should match", key)
	}
}

// Test Delete operation
func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()

	// Delete non-existent key
	existed, err := store.Delete("nonexistent")
	assert.NoError(t, err)
	assert.False(t, existed)

	// Delete existing key
	err = store.Set("key1", "value1", nil)
	require.NoError(t, err)

	existed, err = store.Delete("key1")
	assert.NoError(t, err)
	assert.True(t, existed)

	// Verify key is gone
	value, found := store.Get("key1")
	assert.False(t, found)
	assert.Empty(t, value)

	// Delete empty key
	existed, err = store.Delete("")
	assert.Error(t, err)
	assert.False(t, existed)
}

// Test List operation
func TestMemoryStore_List(t *testing.T) {
	store := NewMemoryStore()

	// List empty store
	result, err := store.List(10)
	assert.NoError(t, err)
	assert.Empty(t, result)

	// Add some data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range testData {
		err := store.Set(key, value, nil)
		require.NoError(t, err)
	}

	// List all
	result, err = store.List(10)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Verify all keys exist
	for key, expectedValue := range testData {
		value, exists := result[key]
		assert.True(t, exists, "Key %s should exist in result", key)
		assert.Equal(t, expectedValue, value, "Value for key %s should match", key)
	}

	// Test limit
	result, err = store.List(2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// Test overwriting existing keys
func TestMemoryStore_Overwrite(t *testing.T) {
	store := NewMemoryStore()

	// Set initial value
	err := store.Set("key1", "value1", nil)
	require.NoError(t, err)

	// Overwrite with new value
	err = store.Set("key1", "value2", nil)
	require.NoError(t, err)

	// Verify new value
	value, found := store.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value2", value)
}

// Test concurrent access
func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				err := store.Set(key, value, nil)
				assert.NoError(t, err)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				store.Get(key) // Value might not exist yet, don't assert
			}
		}(i)
	}

	wg.Wait()

	// Verify some data exists
	result, err := store.List(1000)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

// Add to your existing memory_test.go file

// Test basic TTL expiration
func TestMemoryStore_TTLExpiration(t *testing.T) {
    store := NewMemoryStore()
    
    // Set key with 1 second TTL
    ttl := int64(1)
    err := store.Set("ttl_key", "ttl_value", &ttl)
    require.NoError(t, err)
    
    // Should exist immediately
    value, found := store.Get("ttl_key")
    assert.True(t, found)
    assert.Equal(t, "ttl_value", value)
    
    // Wait for expiration
    time.Sleep(1100 * time.Millisecond)
    
    // Should be expired
    value, found = store.Get("ttl_key")
    assert.False(t, found)
    assert.Empty(t, value)
}

// Test persistent keys (no TTL)
func TestMemoryStore_PersistentKeys(t *testing.T) {
    store := NewMemoryStore()
    
    // Set key without TTL
    err := store.Set("persistent", "value", nil)
    require.NoError(t, err)
    
    // Should persist after reasonable time
    time.Sleep(100 * time.Millisecond)
    value, found := store.Get("persistent")
    assert.True(t, found)
    assert.Equal(t, "value", value)
}

// Test TTL edge cases
func TestMemoryStore_TTLEdgeCases(t *testing.T) {
    store := NewMemoryStore()
    
    // Zero TTL should behave like no TTL
    zeroTTL := int64(0)
    err := store.Set("zero_ttl", "value", &zeroTTL)
    require.NoError(t, err)
    
    value, found := store.Get("zero_ttl")
    assert.True(t, found)
    assert.Equal(t, "value", value)
    
    // Negative TTL should behave like no TTL
    negativeTTL := int64(-1)
    err = store.Set("negative_ttl", "value", &negativeTTL)
    require.NoError(t, err)
    
    value, found = store.Get("negative_ttl")
    assert.True(t, found)
    assert.Equal(t, "value", value)
}

// Test overwriting TTL
func TestMemoryStore_TTLOverwrite(t *testing.T) {
    store := NewMemoryStore()
    
    // Set key with TTL
    ttl := int64(60)
    err := store.Set("key1", "value1", &ttl)
    require.NoError(t, err)
    
    // Overwrite with no TTL
    err = store.Set("key1", "value2", nil)
    require.NoError(t, err)
    
    // Should be persistent now
    value, found := store.Get("key1")
    assert.True(t, found)
    assert.Equal(t, "value2", value)
}

// Test List with expired keys
func TestMemoryStore_ListWithExpiredKeys(t *testing.T) {
    store := NewMemoryStore()
    
    // Set persistent and TTL keys
    err := store.Set("persistent", "value1", nil)
    require.NoError(t, err)
    
    ttl := int64(1)
    err = store.Set("short_ttl", "value2", &ttl)
    require.NoError(t, err)
    
    // Both should appear initially
    result, err := store.List(10)
    assert.NoError(t, err)
    assert.Len(t, result, 2)
    
    // Wait for TTL to expire
    time.Sleep(1100 * time.Millisecond)
    
    // List should only show persistent key
    result, err = store.List(10)
    assert.NoError(t, err)
    assert.Len(t, result, 1)
    assert.Equal(t, "value1", result["persistent"])
}

// Test Delete with TTL keys
func TestMemoryStore_DeleteTTLKeys(t *testing.T) {
    store := NewMemoryStore()
    
    // Set key with TTL
    ttl := int64(60)
    err := store.Set("ttl_key", "value", &ttl)
    require.NoError(t, err)
    
    // Delete should work
    existed, err := store.Delete("ttl_key")
    assert.NoError(t, err)
    assert.True(t, existed)
    
    // Key should be gone
    value, found := store.Get("ttl_key")
    assert.False(t, found)
    assert.Empty(t, value)
}

// Helper function for creating int64 pointers
func int64Ptr(i int64) *int64 {
	return &i
}
