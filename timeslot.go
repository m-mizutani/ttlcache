package ttlcache

type timeSlot[K comparable, V any] struct {
	root node[K, V]
}

func newTimeSlot[K comparable, V any]() *timeSlot[K, V] {
	return &timeSlot[K, V]{}
}

func (x *Cache[K, V]) lookupTimeSlot(t tick) *timeSlot[K, V] {
	if t >= x.cfg.ttl {
		panic("t must be less than TTL")
	}

	idx := (x.current + t) % tick(len(x.timeSlots))
	return x.timeSlots[idx]
}
