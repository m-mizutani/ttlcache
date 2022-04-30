package main

import (
	"fmt"
	"time"

	"github.com/m-mizutani/ttlcache"
)

type myItem struct {
	value int
}

func main() {
	cache := ttlcache.New[string, *myItem]()
	cache.SetHook(func(v *myItem) uint64 {
		fmt.Println("expired!", v)
		return 0
	})

	cache.Set("my_key", &myItem{value: 5}, 4)

	ticker := time.NewTicker(time.Second)
	someQueue := make(chan bool)

	for {
		select {
		case <-someQueue:
			someTask()

		case t := <-ticker.C:
			fmt.Println("Tick at", t)
			cache.Elapse(1)
		}
	}
	// .....
}

func someTask() {}
