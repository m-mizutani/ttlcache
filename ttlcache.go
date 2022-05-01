package ttlcache

import (
	"sync"
)

type (
	tick uint64
)

// CacheTable is TTL cache table. It needs to be created by `New` function.
type CacheTable[K comparable, V any] struct {
	current     tick
	bucketMutex sync.RWMutex
	hookMutex   sync.RWMutex

	bucket    map[K]*node[K, V]
	hooks     map[HookID]Hook[V]
	timeTable *timeTable[K, V]

	baseHookID HookID

	cfg *config
}

type config struct {
	extendByGet bool
	noOverwrite bool
}

type Option func(cfg *config)

// WithExtendByGet enables auto TTL extend when accessing the item by `Get`
func WithExtendByGet() Option {
	return func(cfg *config) {
		cfg.extendByGet = true
	}
}

// WithNoOverwrite stops to overwrite value by `Set` if key already exists.
func WithNoOverwrite() Option {
	return func(cfg *config) {
		cfg.noOverwrite = true
	}
}

// New creates a new `Cache` instance with types K and V. K is for key and V is for value.
func New[K comparable, V any](options ...Option) *CacheTable[K, V] {
	cfg := &config{}

	for _, opt := range options {
		opt(cfg)
	}

	cache := &CacheTable[K, V]{
		cfg:       cfg,
		bucket:    make(map[K]*node[K, V]),
		hooks:     make(map[HookID]Hook[V]),
		timeTable: newTimeTable[K, V](),
	}

	return cache
}

// Set inserts a value to cache with `ttl`. If key already exists, value will be overwritten by default. It always returns `true`, but it returns `false` if WithNoOverwrite enabled and the key already exists.
func (x *CacheTable[K, V]) Set(key K, value V, ttl uint64) bool {
	x.bucketMutex.Lock()
	defer x.bucketMutex.Unlock()

	ttlTick := tick(ttl)

	if n, ok := x.bucket[key]; ok {
		if x.cfg.noOverwrite {
			return false
		}

		n.value = value
		n.last = x.current
		n.ttl = ttlTick
	} else {
		n := &node[K, V]{
			key:   key,
			value: value,
			last:  x.current,
			ttl:   ttlTick,
		}
		x.bucket[key] = n
		slot := x.timeTable.GetOrCreate(ttlTick + x.current)

		n.link = slot.root.link
		slot.root.link = n
	}

	return true
}

// Get looks up `key` from cache and return the value if the `key` exists. If key does not exist, it returns empty value of V (actually `n` of `var n V` will be returned)
func (x *CacheTable[K, V]) Get(key K) V {
	x.bucketMutex.RLock()
	defer x.bucketMutex.RUnlock()

	n, ok := x.bucket[key]
	if !ok {
		var null V
		return null
	}

	if x.cfg.extendByGet {
		n.last = x.current
	}

	return n.value
}

// Elapse puts forward time (tick) of cache table. If a value is expired by forwarding tick, it will be removed from cache table.
func (x *CacheTable[K, V]) Elapse(ticks uint64) {
	x.bucketMutex.RLock()
	defer x.bucketMutex.RUnlock()

	for i := tick(0); i < tick(ticks); i++ {
		x.current++

		slot := x.timeTable.Get(x.current)
		if slot == nil {
			continue
		}

		for {
			n := slot.root.pop()
			if n == nil {
				break
			}

			if x.current <= n.last {
				panic("last accessed tick must be less than updated current")
			}

			var extends []tick

			diff := x.current - n.last
			if diff < n.ttl {
				extends = append(extends, n.ttl-diff)
			} else {
				extends = append(extends, x.runHook(n.value)...)
			}

			var maxExtend tick
			for i := range extends {
				if maxExtend < extends[i] {
					maxExtend = extends[i]
				}
			}

			if maxExtend > 0 {
				x.timeTable.GetOrCreate(x.current + maxExtend).root.push(n)
			} else {
				delete(x.bucket, n.key)
			}
		}
	}
}
