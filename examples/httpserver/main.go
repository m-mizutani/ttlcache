package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/m-mizutani/ttlcache"
)

func main() {
	cache := ttlcache.New[string, string]()
	cache.SetHook(func(v string) uint64 {
		fmt.Println("expired", v)
		return 0
	})

	var lastTick uint64
	startTime := time.Now()

	elapseTick := func() {
		currentTime := time.Now()
		fmt.Println(currentTime)

		delta := currentTime.Sub(startTime)
		tick := uint64(delta / time.Second)
		if tick > lastTick {
			fmt.Println("elapse tick", tick-lastTick)
			cache.Elapse(tick - lastTick)
			lastTick = tick
		}
	}

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		elapseTick()
		fmt.Println("last visitor:", cache.Get("last_visitor"))
		cache.Set("last_visitor", r.RemoteAddr, 4)

		w.Write([]byte("Hello!"))
	})

	http.ListenAndServe("localhost:8000", nil)
}
