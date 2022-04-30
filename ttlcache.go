package ttlcache

import (
	"sync"
)

type (
	tick uint64
)

// Cache is TTL cache table. It needs to be created by `New` function.
type Cache[K comparable, V any] struct {
	current     tick
	bucketMutex sync.RWMutex
	hookMutex   sync.RWMutex

	bucket    map[K]*node[K, V]
	timeSlots []*timeSlot[K, V]
	hooks     map[HookID]Hook[V]

	baseHookID HookID

	cfg *config
}

type config struct {
	noExtend    bool
	noOverwrite bool
	ttl         tick
}

type Option func(cfg *config)

// WithTTL changes time (tick) to live. If you set 4, Elapse(4) will expires cache.
func WithTTL(ttl uint64) Option {
	return func(cfg *config) {
		cfg.ttl = tick(ttl)
	}
}

// WithNoExtend stops to extend TTL of value by `Get`
func WithNoExtend() Option {
	return func(cfg *config) {
		cfg.noExtend = true
	}
}

// WithNoOverwrite stops to overwrite value by `Set` if key already exists.
func WithNoOverwrite() Option {
	return func(cfg *config) {
		cfg.noOverwrite = true
	}
}

// New creates a new `Cache` instance with types K and V. K is for key and V is for value.
func New[K comparable, V any](options ...Option) *Cache[K, V] {
	cfg := &config{
		ttl: 300,
	}

	for _, opt := range options {
		opt(cfg)
	}

	cache := &Cache[K, V]{
		cfg:       cfg,
		bucket:    make(map[K]*node[K, V]),
		timeSlots: make([]*timeSlot[K, V], cfg.ttl),
		hooks:     make(map[HookID]Hook[V]),
	}
	for i := range cache.timeSlots {
		cache.timeSlots[i] = newTimeSlot[K, V]()
	}

	return cache
}

// Set inserts a value to cache with predefined TTL. If key already exists, value will be overwritten by default. It always returns `true`, but it returns `false` if WithNoOverwrite enabled and the key already exists.
func (x *Cache[K, V]) Set(key K, value V) bool {
	x.bucketMutex.Lock()
	defer x.bucketMutex.Unlock()

	if n, ok := x.bucket[key]; ok {
		if x.cfg.noOverwrite {
			return false
		}

		n.value = value
		n.last = x.current
	} else {
		n := &node[K, V]{
			key:   key,
			value: value,
			last:  x.current,
		}
		x.bucket[key] = n
		slot := x.lookupTimeSlot(0)

		n.link = slot.root.link
		slot.root.link = n
	}

	return true
}

// Get looks up `key` from cache and return the value if the `key` exists. If key does not exist, it returns empty value of V (actually `n` of `var n V` will be returned)
func (x *Cache[K, V]) Get(key K) V {
	x.bucketMutex.RLock()
	defer x.bucketMutex.RUnlock()

	n, ok := x.bucket[key]
	if !ok {
		var null V
		return null
	}

	if !x.cfg.noExtend {
		n.last = x.current
	}

	return n.value
}

// Elapse puts forward time (tick) of cache table. If a value is expired by forwarding tick, it will be removed from cache table.
func (x *Cache[K, V]) Elapse(ticks uint64) {
	x.bucketMutex.RLock()
	defer x.bucketMutex.RUnlock()

	for i := tick(0); i < tick(ticks); i++ {
		x.current++

		slot := x.lookupTimeSlot(0)
		for {
			n := slot.root.pop()
			if n == nil {
				break
			}

			if x.current <= n.last {
				panic("last accessed tick must be less than updated current")
			}

			diff := x.current - n.last
			if diff >= x.cfg.ttl {
				delete(x.bucket, n.key)
				x.runHook(n.value)
			} else {
				if x.cfg.ttl < diff {
					panic("x.current - n.last must be less than TTL")
				}
				x.lookupTimeSlot(x.cfg.ttl - diff).root.push(n)
			}
		}
	}
}
