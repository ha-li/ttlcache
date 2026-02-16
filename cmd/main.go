package main

import (
	"fmt"
	"time"
	"ttlcache/cmd/cache"
)

func main() {
	keys := []string{"apple", "orange", "bakeyr"}

	cache := cache.NewTtl(keys, cache.WithTTL(35))
	for i := 0; i < 10; i++ {
		for _, key := range keys {
			v, _ := cache.Get(key)
			fmt.Printf("found %s, value=%s\n", key, v)
		}
		fmt.Println("--------------------------------")
		time.Sleep(8 * time.Second)
	}

	cache.Stop()
	time.Sleep(1 * time.Millisecond) // Give the janitor a moment to print "stopping janitor"
	// fmt.Println("Hello World")
}
