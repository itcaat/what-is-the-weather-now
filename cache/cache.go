package cache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	Data   string
	Expiry time.Time
}

type Cache struct {
	data  map[string]CacheEntry
	mutex sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]CacheEntry),
	}
}

func (c *Cache) Get(key string) (string, bool, time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entry, found := c.data[key]; found && time.Now().Before(entry.Expiry) {
		return entry.Data, true, time.Until(entry.Expiry)
	}
	return "", false, 0
}

func (c *Cache) Set(key, value string, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = CacheEntry{
		Data:   value,
		Expiry: time.Now().Add(ttl),
	}
}
