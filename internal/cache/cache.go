// Package cache provides a simple in-memory TTL cache backed by sync.Map.
// It requires no external dependencies and is safe for concurrent use.
package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

var store sync.Map

// Set stores value under key with the given TTL.
func Set(key string, value any, ttl time.Duration) {
	store.Store(key, entry{value: value, expiresAt: time.Now().Add(ttl)})
}

// Get retrieves a value. Returns (value, true) on hit, (nil, false) on miss or expiry.
func Get(key string) (any, bool) {
	v, ok := store.Load(key)
	if !ok {
		return nil, false
	}
	e := v.(entry)
	if time.Now().After(e.expiresAt) {
		store.Delete(key)
		return nil, false
	}
	return e.value, true
}

// Flush removes all entries (useful for testing).
func Flush() {
	store.Range(func(k, _ any) bool {
		store.Delete(k)
		return true
	})
}
