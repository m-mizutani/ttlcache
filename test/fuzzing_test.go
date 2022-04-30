package ttlcache_test

import (
	"testing"

	"github.com/m-mizutani/ttlcache"
	"github.com/stretchr/testify/assert"
)

func FuzzSetAndGet(f *testing.F) {
	f.Add("abc123", "blue")
	f.Add("abc345", "red")
	cache := ttlcache.New[string, string](
		ttlcache.WithTTL(300),
	)
	seq := 0
	f.Fuzz(func(t *testing.T, key, value string) {
		seq++
		assert.True(t, cache.Set(key, value))
		assert.Equal(t, value, cache.Get(key))
		if seq%100 == 0 {
			cache.Elapse(1)
		}
	})
}
