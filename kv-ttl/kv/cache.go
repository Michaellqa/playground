package kv

import (
	"sync"
	"time"
)

type ttlBox struct {
	CreatedAt time.Time
	Expired   *time.Time
	Value     T
}

type T struct {
	V string
}

type Cache struct {
	mu     sync.RWMutex
	values map[string]ttlBox
}

func NewCache() *Cache {
	c := &Cache{
		mu:     sync.RWMutex{},
		values: make(map[string]ttlBox),
	}
	c.startCleaner(time.Second)
	return c
}

func (c *Cache) startCleaner(delta time.Duration) {
	tick := time.Tick(delta)
	go func() {
		for range tick {
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.values {
				if v.Expired != nil && now.After(*v.Expired) {
					delete(c.values, k)
				}
			}
			c.mu.Unlock()
		}
	}()
}

// --- crud ---

func (c *Cache) Add(key string, value T) bool {
	return c.add(key, value, nil)
}

func (c *Cache) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.values[key]
	return value.Value, ok
}

func (c *Cache) GetAll() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	results := make([]T, 0, len(c.values))
	for _, b := range c.values {
		results = append(results, b.Value)
	}
	return results
}

func (c *Cache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.values, key)
}

// --- ttl ---

func (c *Cache) AddWithTtl(key string, value T, ttl time.Duration) bool {
	expired := time.Now().Add(ttl)
	return c.add(key, value, &expired)
}

func (c *Cache) GetTtl(key string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.values[key]
	if !ok {
		return 0, false
	}
	return time.Now().Sub(value.CreatedAt), true
}

func (c *Cache) SetTtl(key string, ttl *time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.values[key]
	if !ok {
		return false
	}
	value.Expired = ttl
	c.values[key] = value
	return true
}

// --- private ---

func (c *Cache) add(key string, value T, ttl *time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.values[key]; ok {
		return false
	}
	c.values[key] = ttlBox{
		CreatedAt: time.Now(),
		Expired:   ttl,
		Value:     value,
	}
	return true
}
