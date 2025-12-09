package storage

import (
	"sync"
	"time"
)

type CacheItem struct {
	value      string
	expiration int64
}

type Cache struct {
	data  map[string]CacheItem
	mutex sync.RWMutex
}

func NewCache() *Cache {
	c := &Cache{
		data: make(map[string]CacheItem),
	}
	go c.CleanUp()

	return c
}

func (c *Cache) Set(key string, value string, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = CacheItem{
		value:      value,
		expiration: time.Now().Add(ttl).UnixNano(),
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mutex.RLock()
	item, ok := c.data[key]
	c.mutex.RUnlock()

	if !ok {
		return "", false
	}

	if time.Now().UnixNano() > item.expiration {
		// Item expired, remove it
		c.mutex.Lock()
		delete(c.data, key)
		c.mutex.Unlock()
		return "", false
	}

	return item.value, true
}

func (c *Cache) CleanUp() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now().UnixNano()

		for key, item := range c.data {
			if now > item.expiration {
				delete(c.data, key)
			}
		}

		c.mutex.Unlock()
	}
}
