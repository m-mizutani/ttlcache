package ttlcache_test

import (
	"testing"

	"github.com/m-mizutani/ttlcache"
	"github.com/stretchr/testify/assert"
)

func FuzzSetAndGet(f *testing.F) {
	f.Add("abc123", "blue", uint64(3))
	f.Add("abc345", "red", uint64(2))
	cache := ttlcache.New[string, string]()

	seq := 0
	f.Fuzz(func(t *testing.T, key, value string, ttl uint64) {
		seq++
		assert.True(t, cache.Set(key, value, ttl))
		assert.Equal(t, value, cache.Get(key))
		if seq%100 == 0 {
			cache.Elapse(1)
		}
	})
}
