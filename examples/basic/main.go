package main

import (
	"fmt"

	"github.com/m-mizutani/ttlcache"
)

type myItem struct {
	value int
}

func main() {
	cache := ttlcache.New[string, *myItem]()

	// Set own struct with TTL:10
	cache.Set("my_key", &myItem{value: 5}, 10)

	// Elapses 9 ticks, but cache item is still not expired
	cache.Elapse(9)

	// my_key => &{5}
	fmt.Println("my_key =>", cache.Get("my_key"))

	// Elapses 1 ticks, then cache item was expired because of reaching out to TTL.
	cache.Elapse(1)

	// my_key => <nil>
	fmt.Println("my_key =>", cache.Get("my_key"))
}
