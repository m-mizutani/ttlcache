package ttlcache_test

import (
	"testing"

	"github.com/m-mizutani/ttlcache"

	"github.com/stretchr/testify/assert"
)

func TestExpires(t *testing.T) {
	t.Run("not expire before TTL", func(t *testing.T) {
		cache := ttlcache.New[string, string]()
		key := "one"
		cache.Set(key, "void", 4)
		cache.Elapse(3)
		assert.Equal(t, "void", cache.Get(key))
	})

	t.Run("expire cache", func(t *testing.T) {
		cache := ttlcache.New[string, string]()
		cache.Set("one", "void", 4)
		cache.Elapse(1)
		cache.Set("five", "blue", 5)
		cache.Elapse(3)
		assert.Equal(t, "", cache.Get("one")) // expired
		assert.Equal(t, "blue", cache.Get("five"))

		cache.Elapse(1)
		assert.Equal(t, "blue", cache.Get("five"))
		cache.Elapse(1)
		assert.Equal(t, "", cache.Get("five")) // expired

	})
}

func TestAutoExtend(t *testing.T) {
	t.Run("not expire with access", func(t *testing.T) {
		cache := ttlcache.New[string, string](
			ttlcache.WithExtendByGet(),
		)

		key := "one"
		cache.Set(key, "void", 4)
		cache.Elapse(1)
		assert.Equal(t, "void", cache.Get(key))
		cache.Elapse(3)
		assert.Equal(t, "void", cache.Get(key)) // not expired because of extending by Get

		cache.Elapse(3)
		assert.Equal(t, "void", cache.Get(key)) // also
		cache.Elapse(4)
		assert.Equal(t, "", cache.Get(key)) // expired over TTL ticks
	})
}

func TestHook(t *testing.T) {
	t.Run("run hook by expire cache", func(t *testing.T) {
		var called int
		cache := ttlcache.New[string, string]()
		cache.SetHook(func(v string) uint64 {
			called++
			assert.Equal(t, "void", v)
			return 0
		})
		key := "one"
		cache.Set(key, "void", 4)
		cache.Elapse(4)
		assert.Equal(t, "", cache.Get(key))
		assert.Equal(t, 1, called)
	})

	t.Run("not run deleted hook", func(t *testing.T) {
		cache := ttlcache.New[string, string]()
		id := cache.SetHook(func(v string) uint64 {
			assert.Fail(t, "hook should not be called")
			return 0
		})
		key := "one"
		cache.Set(key, "void", 4)
		cache.Elapse(2)
		assert.True(t, cache.DelHook(id))
		cache.Elapse(2)
		assert.Equal(t, "", cache.Get(key))
	})

	t.Run("extend by hook", func(t *testing.T) {
		var called int
		cache := ttlcache.New[string, string]()
		cache.SetHook(func(v string) uint64 {
			called++
			if called == 1 {
				return 1
			}
			return 0
		})

		key := "one"
		cache.Set(key, "void", 4)
		cache.Elapse(4)
		assert.Equal(t, "void", cache.Get(key)) // live more 1 tick
		cache.Elapse(1)
		assert.Equal(t, "", cache.Get(key)) // extending does not work in 2nd time
		assert.Equal(t, 2, called)
	})

	t.Run("extend with largest tick", func(t *testing.T) {
		cache := ttlcache.New[string, string]()
		var notExtend bool
		cache.SetHook(func(v string) uint64 {
			if notExtend {
				return 0
			}
			return 1
		})
		cache.SetHook(func(v string) uint64 {
			if notExtend {
				return 0
			}
			return 5
		})
		cache.SetHook(func(v string) uint64 {
			if notExtend {
				return 0
			}
			return 3
		})

		key := "one"
		cache.Set(key, "void", 1)
		cache.Elapse(1)
		assert.Equal(t, "void", cache.Get(key)) // extended

		notExtend = true // enable stopper

		cache.Elapse(1)                         // +1
		assert.Equal(t, "void", cache.Get(key)) // not expired
		cache.Elapse(1)                         // +2
		assert.Equal(t, "void", cache.Get(key)) // not expired
		cache.Elapse(1)                         // +3
		assert.Equal(t, "void", cache.Get(key)) // not expired
		cache.Elapse(1)                         // +4
		assert.Equal(t, "void", cache.Get(key)) // not expired
		cache.Elapse(1)                         // +5
		assert.Equal(t, "", cache.Get(key))     // finally expired. TTL is extended with 5 tick
	})
}

func TestDeleteNode(t *testing.T) {
	cache := ttlcache.New[string, string]()
	cache.SetHook(func(v string) uint64 {
		assert.Equal(t, "will be deleted by Expire", v)
		return 0
	})

	cache.Set("a", "will be deleted by Delete", 4)
	cache.Set("b", "will be deleted by Expire", 4)
	cache.Elapse(2)
	assert.True(t, cache.Delete("a"))
	assert.False(t, cache.Delete("a")) // 2nd delete will be failed, but only returned false
	assert.Empty(t, cache.Get("a"))

	cache.Elapse(2) // b also deleted by expiring
	assert.Empty(t, cache.Get("b"))
}
