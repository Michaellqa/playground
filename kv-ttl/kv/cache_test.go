package kv

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestCrazyCache(t *testing.T) {
	c := NewCache()
	time.Sleep(2 * time.Second)

	v := rand.Intn(10000)
	key := strconv.Itoa(v)

	go func() {
		c.AddWithTtl(key, T{key}, time.Second)
	}()
}
