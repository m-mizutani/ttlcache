package ttlcache

type node[K comparable, V any] struct {
	key   K
	value V
	ttl   tick
	last  tick
	link  *node[K, V]

	deleted bool
}

func (x *node[K, V]) push(n *node[K, V]) {
	n.link = x.link
	x.link = n
}

func (x *node[K, V]) pop() *node[K, V] {
	n := x.link
	if n == nil {
		return nil
	}
	x.link = n.link
	n.link = nil
	return n
}
