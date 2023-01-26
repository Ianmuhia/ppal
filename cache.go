package main

import (
	"log"
	"sync"
)

// Cache holds the structure for
// holding a connected clients.
type Cache struct {
	data map[string]*Client
	lock *sync.RWMutex
}

// NewCache initializes the cache.
func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*Client),
		lock: &sync.RWMutex{},
	}
}

// Get retrieves the client from cache.
func (c *Cache) Get(key string) (*Client, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v, ok := c.data[key]
	if !ok {
		return nil, false
	}

	return v, true
}

// Set Adds the client to cache.
func (c *Cache) Set(val *Client, key string) {
	defer c.lock.Unlock()
	if c.lock.TryLock() {
		c.data[key] = val
		log.Printf("wrote key: [%v]  value: [%v]  to cache", val.id, val.rank)
	}
} // Del Deletes the client to cache.
func (c *Cache) Del(key string) string {
	defer c.lock.Unlock()
	if c.lock.TryLock() {
		delete(c.data, key)
		log.Printf("deleted key: [%v]  to cache", key)
	}
	return key
}
