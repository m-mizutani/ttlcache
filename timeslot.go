package ttlcache

type timeTable[K comparable, V any] struct {
	slots map[tick]*timeSlot[K, V]
}

func newTimeTable[K comparable, V any]() *timeTable[K, V] {
	return &timeTable[K, V]{
		slots: make(map[tick]*timeSlot[K, V]),
	}
}

func (x *timeTable[K, V]) Get(t tick) *timeSlot[K, V] {
	if slot, ok := x.slots[t]; ok {
		return slot
	}

	return nil
}

func (x *timeTable[K, V]) GetOrCreate(t tick) *timeSlot[K, V] {
	if slot := x.Get(t); slot != nil {
		return slot
	}

	slot := newTimeSlot[K, V]()
	x.slots[t] = slot
	return slot
}

func (x *timeTable[K, V]) Purge(t tick) {
	if _, ok := x.slots[t]; ok {
		delete(x.slots, t)
	}
}

type timeSlot[K comparable, V any] struct {
	root node[K, V]
}

func newTimeSlot[K comparable, V any]() *timeSlot[K, V] {
	return &timeSlot[K, V]{}
}
