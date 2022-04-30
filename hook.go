package ttlcache

type (
	HookID uint64

	// Hook function is triggered by expiring a cache item. By returning more than 0, TTL of the cache item will be extended. For example, the cache item has more 5 tick TTL by returning 5. If multiple hooks return more than 0, the cache item will be extended with a largest value in returned. NOTE: DO NOT access the cache table in Hook function to avoid DEAD LOCK.
	Hook[V any] func(V) uint64
)

// SetHook appends a hook function to be triggered by expring a cache item. SetHook returns ID or appended hook and it can be used to remove the hook.
func (x *CacheTable[K, V]) SetHook(h Hook[V]) HookID {
	x.hookMutex.Lock()
	defer x.hookMutex.Unlock()

	x.baseHookID++
	id := x.baseHookID

	x.hooks[id] = h
	return id
}

// DelHook removes appended hook function with ID provided by SetHook.
func (x *CacheTable[K, V]) DelHook(id HookID) bool {
	x.hookMutex.Lock()
	defer x.hookMutex.Unlock()

	if _, ok := x.hooks[id]; !ok {
		return false
	}
	delete(x.hooks, id)
	return true
}

func (x *CacheTable[K, V]) runHook(value V) []tick {
	x.hookMutex.RLock()
	defer x.hookMutex.RUnlock()

	var resp []tick
	for _, hook := range x.hooks {
		resp = append(resp, tick(hook(value)))
	}

	return resp
}
