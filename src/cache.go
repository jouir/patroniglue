package main

import (
	"sync"
	"time"
)

// Cache defines an interface to implement caching structure
type Cache interface {
	Startup()
	Get(string) (interface{}, error)
	Set(string, interface{}) error
}

// MemoryCache caches data in an in-memory key-value store with a ttl in seconds
type MemoryCache struct {
	mutex       sync.Mutex
	enabled     bool
	datastore   map[string]interface{}
	expireStore map[string]time.Time
	ttl         float64
	interval    float64
}

// NewCache creates a Cache instance
func NewCache(config CacheConfig) (Cache, error) {
	enabled := false
	if config.TTL > 0 {
		enabled = true
	}
	if config.Interval == 0 {
		config.Interval = 0.25
	}
	return &MemoryCache{
		ttl:      config.TTL,
		interval: config.Interval,
		enabled:  enabled,
	}, nil
}

// Startup starts cache management threads
func (c *MemoryCache) Startup() {
	Debug("starting memory cache")
	c.datastore = make(map[string]interface{})
	c.expireStore = make(map[string]time.Time)
	go c.expireThread()
}

// expireThread flushes datastore every ttl seconds
func (c *MemoryCache) expireThread() {
	Debug("starting cache expire thread")
	if c.enabled {
		for {
			c.mutex.Lock()
			for key := range c.expireStore {
				if time.Since(c.expireStore[key]).Seconds() > c.ttl {
					Debug("deleting key '%s' from cache", key)
					delete(c.datastore, key)
					delete(c.expireStore, key)
				}
			}
			c.mutex.Unlock()
			time.Sleep(time.Duration(c.interval*1000) * time.Millisecond)
		}
	}
	Debug("ending cache expire thread")
}

// Get a value from cache datastore
func (c *MemoryCache) Get(key string) (interface{}, error) {
	if !c.enabled {
		return nil, nil
	}
	value, ok := c.datastore[key]
	if ok {
		Debug("value for key '%s' found in cache", key)
		return value, nil
	}
	Debug("value for key '%s' not found in cache", key)
	return nil, nil
}

// Set a value into cache datastore with a key
func (c *MemoryCache) Set(key string, value interface{}) error {
	if !c.enabled {
		return nil
	}
	Debug("setting value for key '%s' in cache", key)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.datastore[key] = value
	c.expireStore[key] = time.Now()
	return nil
}
