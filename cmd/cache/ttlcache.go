package cache

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type item struct {
	value string
	ttl   int64
}

const (
	defaultTTL int64 = 10 // seconds
)

type ttlCache struct {
	mu         sync.RWMutex
	keys       []string
	store      map[string]item
	lastUpdate int64
}

type Cache interface {
	Get(key string) (string, bool)
}

func NewTtl(keysToSave []string) Cache {
	c := ttlCache{
		keys:  keysToSave,
		store: make(map[string]item),
	}

	store, lastUpdate := c.load()
	c.store = store
	c.lastUpdate = lastUpdate

	// start the janitor
	c.janitorWithRefresh()
	return &c
}

func (c *ttlCache) load() (map[string]item, int64) {
	nc := make(map[string]item)
	for _, k := range c.keys {
		nv := c.get()
		i := item{nv, defaultTTL}
		nc[k] = i
	}
	return nc, time.Now().Unix()
}

func (c *ttlCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	i, found := c.store[key]
	if !found {
		return "", false
	}
	return i.value, true
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (c *ttlCache) get() string {
	return generateRandomString(10)
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

func (c *ttlCache) janitorWithRefresh() {
	fmt.Println("starting janitor")
	go func() {
		for range time.Tick(60 * time.Second) {
			fmt.Println("janitor running")

			c.mu.RLock()
			elapsed := time.Now().Unix() - c.lastUpdate
			c.mu.RUnlock()

			if elapsed > defaultTTL {
				fmt.Println("refreshing cache")
				newCache, lastUpdate := c.load()

				// swap the cache
				c.mu.Lock()
				c.store = newCache
				c.lastUpdate = lastUpdate
				c.mu.Unlock()
			}
			fmt.Println("janitor finished")
		}
	}()
}
