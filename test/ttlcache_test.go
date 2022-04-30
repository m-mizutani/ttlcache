package ttlcache_test

import (
	"testing"

	"github.com/m-mizutani/ttlcache"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	cache := ttlcache.New[string](
		ttlcache.WithTTL(4),
	)

	key := []byte("one")
	cache.Set(key, "void")
	cache.Elapse(1)
	assert.Equal(t, "void", cache.Get(key))
	cache.Elapse(1)
	assert.Equal(t, "void", cache.Get(key))
	cache.Elapse(1) // expired
	assert.Equal(t, "", cache.Get(key))
}

func TestHook(t *testing.T) {
	var called int
	hook := func(v string) {
		called++
	}
	cache := ttlcache.New[string](
		ttlcache.WithTTL(4),
	)

	id := cache.SetHook(hook)

	key := []byte("one")
	cache.Set(key, "void")
	cache.Elapse(3) // expired
	assert.Equal(t, "", cache.Get(key))

	assert.Equal(t, 1, called)

	assert.True(t, cache.DelHook(id))

	cache.Set(key, "void")
	cache.Elapse(3) // expired
	assert.Equal(t, "", cache.Get(key))
	assert.Equal(t, 1, called)
}
