package ttlcache_test

import (
	"testing"

	"github.com/m-mizutani/ttlcache"

	"github.com/stretchr/testify/assert"
)

func TestExpires(t *testing.T) {
	t.Run("not expire before TTL", func(t *testing.T) {
		cache := ttlcache.New[string, string](
			ttlcache.WithTTL(4),
		)
		key := "one"
		cache.Set(key, "void")
		cache.Elapse(3)
		assert.Equal(t, "void", cache.Get(key))
	})

	t.Run("expire cache", func(t *testing.T) {
		cache := ttlcache.New[string, string](
			ttlcache.WithTTL(4),
		)
		cache.Set("one", "void")
		cache.Elapse(1)
		cache.Set("five", "blue")
		cache.Elapse(3)
		assert.Equal(t, "", cache.Get("one"))
		assert.Equal(t, "blue", cache.Get("five"))
	})

	t.Run("not expire with access", func(t *testing.T) {
		cache := ttlcache.New[string, string](
			ttlcache.WithTTL(4),
		)

		key := "one"
		cache.Set(key, "void")
		cache.Elapse(1)
		assert.Equal(t, "void", cache.Get(key))
		cache.Elapse(1)
		assert.Equal(t, "void", cache.Get(key))
		cache.Elapse(1)
		assert.Equal(t, "void", cache.Get(key))
		cache.Elapse(1) // not expired
		assert.Equal(t, "void", cache.Get(key))
		cache.Elapse(4) // expires
		assert.Equal(t, "", cache.Get(key))
	})

	t.Run("not expire with access case2", func(t *testing.T) {
		cache := ttlcache.New[string, string](
			ttlcache.WithTTL(10),
		)

		key := "one"
		cache.Set(key, "void")
		cache.Elapse(1)
		assert.Equal(t, "void", cache.Get(key))
		cache.Elapse(10)
		assert.Equal(t, "", cache.Get(key))
	})
}

func TestNotExtend(t *testing.T) {
	t.Run("not expire with access", func(t *testing.T) {
		cache := ttlcache.New[string, string](
			ttlcache.WithTTL(4),
			ttlcache.WithNoExtend(),
		)

		key := "one"
		cache.Set(key, "void")
		cache.Elapse(1)
		assert.Equal(t, "void", cache.Get(key))
		cache.Elapse(3)
		assert.Equal(t, "", cache.Get(key)) // expires even if accessed
	})
}

func TestHook(t *testing.T) {
	t.Run("run hook by expire cache", func(t *testing.T) {
		var called int
		cache := ttlcache.New[string, string](
			ttlcache.WithTTL(4),
		)
		cache.SetHook(func(v string) {
			called++
			assert.Equal(t, "void", v)
		})
		key := "one"
		cache.Set(key, "void")
		cache.Elapse(4)
		assert.Equal(t, "", cache.Get(key))
		assert.Equal(t, 1, called)
	})

	t.Run("not run deleted hook", func(t *testing.T) {
		cache := ttlcache.New[string, string](
			ttlcache.WithTTL(4),
		)
		id := cache.SetHook(func(v string) {
			assert.Fail(t, "hook should not be called")
		})
		key := "one"
		cache.Set(key, "void")
		cache.Elapse(2)
		assert.True(t, cache.DelHook(id))
		cache.Elapse(2)
		assert.Equal(t, "", cache.Get(key))
	})
}
