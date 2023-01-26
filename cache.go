package main

import (
	"log"
	"sync"
)

type Cache struct {
	data map[string]*ConnectedUsers
	lock *sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*ConnectedUsers),
		lock: &sync.RWMutex{},
	}
}

// Get retrieves the token from cache.
func (c *Cache) Get(key string) (*ConnectedUsers, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v, ok := c.data[key]
	if !ok {
		return nil, false
	}

	return v, true
}

// Set Adds the token to cache.
func (c *Cache) Set(val *ConnectedUsers, key string) {
	defer c.lock.Unlock()
	if c.lock.TryLock() {
		c.data[key] = val
		log.Printf("wrote key: [%v]  value: [%v]  to cache", val.id, val.rank)
	}
} // Del Deletes the token to cache.
func (c *Cache) Del(key string) string {
	defer c.lock.Unlock()
	if c.lock.TryLock() {
		delete(c.data, key)
		log.Printf("deleted key: [%v]  to cache", key)
	}
	return key
}
