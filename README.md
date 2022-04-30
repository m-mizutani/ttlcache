# ttlcache [![test](https://github.com/m-mizutani/ttlcache/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/ttlcache/actions/workflows/test.yml) [![pkg-scan](https://github.com/m-mizutani/ttlcache/actions/workflows/trivy.yml/badge.svg)](https://github.com/m-mizutani/ttlcache/actions/workflows/trivy.yml) [![gosec](https://github.com/m-mizutani/ttlcache/actions/workflows/gosec.yml/badge.svg)](https://github.com/m-mizutani/ttlcache/actions/workflows/gosec.yml)

Time to live cache table with Go generics.

```go
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
```

`ttlcache` does not manage real "time", but manages time to live of cache items by `tick`. You can set time to live as `tick` with `Set()`. The item will be expired by forwarding "tick" with `Elapse()`.

If you want to tick as real 1 second, `time.Ticker` can be used to elapse `tick` of `ttlcache` as following example.

```go
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
```

That code outputs like following.

```bash
Tick at 2022-04-30 19:29:52.94181 +0900 JST m=+1.001117668
Tick at 2022-04-30 19:29:53.944413 +0900 JST m=+2.003701334
Tick at 2022-04-30 19:29:54.941845 +0900 JST m=+3.001114334
Tick at 2022-04-30 19:29:55.941862 +0900 JST m=+4.001110876
expired! &{5}
```

Example codes are available in [examples](./examples/).

## Other features

- Type safe by generics
- Thread safe
- Hook function to be triggered with expiration of a cache item
  - Hook function can extend TTL of expired cache item
- Options
  - `WithNoOverwrite`: Stops to overwrite value by `Set` when the key already exists.
  - `WithExtendByGet`: Extends TTL by accessing the cache item with original TTL.

## License

Apache License Version 2.0
