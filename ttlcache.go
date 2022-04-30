package ttlcache

import (
	"bytes"
	"hash/fnv"
	"sync"

	"github.com/google/uuid"
)

type HookID string

type Hook[T any] func(T)

type Cache[T any] struct {
	current   uint64
	tickMutex sync.Mutex
	hookMutex sync.RWMutex

	buckets []*bucket[T]
	timers  []*timer[T]
	hooks   map[HookID]Hook[T]
}

type config struct {
	ttl  uint64
	size uint64
}

type Option func(cfg *config)

func WithSize(size uint64) Option {
	return func(cfg *config) {
		cfg.size = size
	}
}

func WithTTL(ttl uint64) Option {
	return func(cfg *config) {
		cfg.ttl = ttl
	}
}

func New[T any](options ...Option) *Cache[T] {
	cfg := &config{
		ttl:  300,
		size: 1024,
	}

	for _, opt := range options {
		opt(cfg)
	}

	cache := &Cache[T]{
		buckets: make([]*bucket[T], cfg.size),
		timers:  make([]*timer[T], cfg.ttl),
		hooks:   make(map[HookID]Hook[T]),
	}
	for i := range cache.buckets {
		cache.buckets[i] = newBucket[T]()
	}
	for i := range cache.timers {
		cache.timers[i] = newTimer[T]()
	}

	return cache
}

func (x *Cache[T]) Set(key []byte, value T) {
	bucket := x.lookupBucket(key)
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	for p := bucket.root.next; p != nil; p = p.next {
		// overwrite
		if bytes.Equal(key, p.key) {
			p.value = value
			return
		}
	}

	n := &node[T]{
		key:    key[:],
		value:  value,
		bucket: bucket,
	}

	bucket.root.attach(n)
}

func (x *Cache[T]) Get(key []byte) T {
	bucket := x.lookupBucket(key)
	bucket.mutex.RLock()
	defer bucket.mutex.RUnlock()

	for p := bucket.root.next; p != nil; p = p.next {
		if bytes.Equal(key, p.key) {
			return p.value
		}
	}

	var null T
	return null
}

func (x *Cache[T]) SetHook(h Hook[T]) HookID {
	id := HookID(uuid.NewString())

	x.hookMutex.Lock()
	defer x.hookMutex.Unlock()

	x.hooks[id] = h
	return id
}

func (x *Cache[T]) DelHook(id HookID) bool {
	x.hookMutex.Lock()
	defer x.hookMutex.Unlock()

	if _, ok := x.hooks[id]; !ok {
		return false
	}
	delete(x.hooks, id)
	return true
}

func (x *Cache[T]) runHook(v T) {
	x.hookMutex.RLock()
	defer x.hookMutex.RUnlock()

	for _, hook := range x.hooks {
		hook(v)
	}
}

func (x *Cache[T]) Elapse(tick int) {
	for i := 0; i < tick; i++ {
		x.current++
		timer := x.lookupTimer()

		for n := timer.root.link; n != nil; n = n.link {
			n.detach()
		}
	}
}

func (x *Cache[T]) lookupBucket(key []byte) *bucket[T] {
	idx := hash(key) % uint64(len(x.buckets))
	return x.buckets[idx]
}

func (x *Cache[T]) lookupTimer() *timer[T] {
	idx := x.current % uint64(len(x.timers))
	return x.timers[idx]
}

func hash(key []byte) uint64 {
	if len(key) == 0 {
		return 0
	}

	hash := fnv.New64a()
	hash.Write(key)
	return hash.Sum64()
}

type node[T any] struct {
	key        []byte
	prev, next *node[T]
	link       *node[T]
	value      T
	bucket     *bucket[T]
}

func (x *node[T]) attach(n *node[T]) {
	next := x.next
	prev := x

	if next != nil {
		next.prev = n
	}
	prev.next = n

	n.next = next
	n.prev = prev
}

func (x *node[T]) detach() {
	next := x.next
	prev := x.prev
	if next != nil {
		next.prev = prev
	}
	if prev != nil {
		prev.next = next
	}
	x.prev, x.next = nil, nil
}

type bucket[T any] struct {
	mutex sync.RWMutex
	root  node[T]
}

func newBucket[T any]() *bucket[T] {
	return &bucket[T]{}
}

type timer[T any] struct {
	root node[T]
}

func newTimer[T any]() *timer[T] {
	return &timer[T]{}
}
