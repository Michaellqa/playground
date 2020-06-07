package kv

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	config configuration
	mu     sync.RWMutex
	values map[string]ttlBox
}

func NewCache(config Configuration) *Cache {
	c := &Cache{
		mu:     sync.RWMutex{},
		values: make(map[string]ttlBox),
	}
	c.configure(config)
	c.startCleaner(time.Second)
	return c
}

func (c *Cache) configure(config Configuration) {
	c.config = configuration{}

	if config.FileName != "" {
		c.config.fileName = config.FileName
		c.restore()
	} else {
		b := make([]byte, 6)
		if _, err := rand.Read(b); err != nil {
			log.Println(err)
		}
		c.config.fileName = fmt.Sprintf("snapshot_%x.json", b)
	}

	if config.BackupInterval != 0 {
		c.config.backupInterval = config.BackupInterval
		c.startAutoBackup()
	}
}

func (c *Cache) restore() {
	f, err := os.OpenFile(c.config.fileName, os.O_RDONLY, 0666)
	if err != nil {
		log.Printf("%v - initiated a new cache\n", err)
		return
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(&c.values); err != nil {
		log.Printf("%v - initiated a new cache\n", err)
	}
}

// --- background ---

func (c *Cache) startCleaner(delta time.Duration) {
	tick := time.Tick(delta)
	go func() {
		for range tick {
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.values {
				if v.Expired != nil && now.After(*v.Expired) {
					fmt.Printf("deleted: %s %v\n", k, v)
					delete(c.values, k)
				}
			}
			c.mu.Unlock()
		}
	}()
}

func (c *Cache) startAutoBackup() {
	tick := time.Tick(c.config.backupInterval)
	go func() {
		for range tick {
			c.makeSnapshot()
		}
	}()
}

func (c *Cache) makeSnapshot() {
	c.mu.RLock()
	data, err := json.Marshal(c.values)
	c.mu.RUnlock()
	if err != nil {
		log.Println(err)
		return
	}

	f, err := os.OpenFile(c.config.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	log.Printf("file <= %s", data)
	if _, err := f.Write(data); err != nil {
		log.Println(err)
	}
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
