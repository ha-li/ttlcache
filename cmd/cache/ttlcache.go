package cache

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type item struct {
	value string
}

const (
	defaultTTL int64 = 10 // seconds
)

type ttlCache struct {
	mu         sync.RWMutex
	keys       []string
	store      map[string]item
	lastUpdate int64
	ttl        int64
	done       chan struct{}
}

type Cache interface {
	Get(key string) (string, bool)
	Stop()
}

func NewTtl(keysToSave []string, opts ...Option) Cache {
	c := &ttlCache{
		done: make(chan struct{}),
		ttl:  defaultTTL,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.keys = keysToSave
	c.store = make(map[string]item)

	store, lastUpdate := c.load()
	c.store = store
	c.lastUpdate = lastUpdate

	// start the janitor
	c.janitorWithRefresh()
	return c
}

func (c *ttlCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	i, found := c.store[key]
	if !found {
		return "", false
	}
	return i.value, true
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateValue generates a random string of length 10
func (c *ttlCache) generateValue() string {
	return generateRandomString(10)
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

func (c *ttlCache) load() (map[string]item, int64) {
	nc := make(map[string]item)
	for _, k := range c.keys {
		nv := c.generateValue() // generate a random value
		i := item{nv}
		nc[k] = i
	}
	return nc, time.Now().Unix()
}

func (c *ttlCache) Stop() {
	close(c.done)
}

func (c *ttlCache) janitorWithRefresh() {
	// a ticker of ttl/2 secs, means every ttl/2 sec, the ticker gets
	// an event (ie the case: <-ticker.C)
	ticker := time.NewTicker(time.Duration(c.ttl/2) * time.Second)
	fmt.Println("starting janitor")
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-c.done:
				fmt.Println("stopping janitor")
				return
			case <-ticker.C: // is true on each ticker duration

				c.mu.RLock()
				fmt.Println("janitor running")

				// time since last update
				elapsed := time.Now().Unix() - c.lastUpdate
				// refresh every ttl which is 2x ticker.C
				shouldRefresh := elapsed > c.ttl
				c.mu.RUnlock()

				if shouldRefresh {
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
		}
	}()
}
