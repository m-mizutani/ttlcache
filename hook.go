package ttlcache

type (
	HookID uint64

	Hook[V any] func(V)
)

// SetHook appends a hook function to be triggered with removing a value. SetHook returns ID or appended hook and it can be used to remove the hook.
func (x *Cache[K, V]) SetHook(h Hook[V]) HookID {

	x.hookMutex.Lock()
	defer x.hookMutex.Unlock()

	x.baseHookID++
	id := x.baseHookID

	x.hooks[id] = h
	return id
}

// DelHook removes appended hook function with ID provided by SetHook.
func (x *Cache[K, V]) DelHook(id HookID) bool {
	x.hookMutex.Lock()
	defer x.hookMutex.Unlock()

	if _, ok := x.hooks[id]; !ok {
		return false
	}
	delete(x.hooks, id)
	return true
}

func (x *Cache[K, V]) runHook(value V) {
	x.hookMutex.RLock()
	defer x.hookMutex.RUnlock()

	for _, hook := range x.hooks {
		hook(value)
	}
}
