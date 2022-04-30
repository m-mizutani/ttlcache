# ttlcache

Time to live cache table with Go generics.

```go
type myItem struct {
	value int
}

func main() {
	cache := ttlcache.New[string, *myItem](
		ttlcache.WithTTL(10),
	)

    // Set own type
	cache.Set("my_key", &myItem{value: 5})

	// my_key => &{5}
	fmt.Println("my_key =>", cache.Get("my_key"))

	// Elapse 10 ticks expires the item
	cache.Elapse(10)

	// my_key => <nil>
	fmt.Println("my_key =>", cache.Get("my_key"))
}
```
