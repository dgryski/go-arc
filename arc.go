// Package arc implements the Adaptive Replacement Cache
/*

https://www.usenix.org/legacy/events/fast03/tech/full_papers/megiddo/megiddo.pdf

This code is a straight-forward translation of the Python implementation at
http://code.activestate.com/recipes/576532-adaptive-replacement-cache-in-python/
modified to make the O(n) list operations O(1).

It is MIT licensed, like the original Python implementation.

*/
package arc

import "container/list"

// Cache is a type implementing an Adaptive Replacement Cache
type Cache struct {
	data map[string][]byte

	cap  int
	part int

	t1 *clist
	t2 *clist
	b1 *clist
	b2 *clist
}

func newClist() *clist {
	return &clist{
		l:    list.New(),
		keys: make(map[string]*list.Element),
	}
}

type clist struct {
	l    *list.List
	keys map[string]*list.Element
}

func (c *clist) Has(key string) bool {
	_, ok := c.keys[key]
	return ok
}

func (c *clist) Lookup(key string) *list.Element {
	elt := c.keys[key]
	return elt
}

func (c *clist) MoveToFront(elt *list.Element) {
	c.l.MoveToFront(elt)
}

func (c *clist) PushFront(key string) {
	elt := c.l.PushFront(key)
	c.keys[key] = elt
}

func (c *clist) Remove(key string, elt *list.Element) {
	delete(c.keys, key)
	c.l.Remove(elt)
}

func (c *clist) RemoveTail() string {
	elt := c.l.Back()
	c.l.Remove(elt)

	key := elt.Value.(string)
	delete(c.keys, key)

	return key
}

func (c *clist) Len() int {
	return c.l.Len()
}

// New creates an ARC that stores at most size items.
func New(size int) *Cache {
	return &Cache{
		data: make(map[string][]byte),
		cap:  size,
		t1:   newClist(),
		t2:   newClist(),
		b1:   newClist(),
		b2:   newClist(),
	}
}

func (c *Cache) replace(key string) {

	var old string
	if (c.t1.Len() > 0 && c.b2.Has(key) && c.t1.Len() == c.part) || (c.t1.Len() > c.part) {
		old = c.t1.RemoveTail()
		c.b1.PushFront(old)
	} else {
		old = c.t2.RemoveTail()
		c.b2.PushFront(old)
	}

	delete(c.data, old)
}

// Get retrieves a value from the cache. The function f will be called to
// retrieve the value if it is not present in the cache.
func (c *Cache) Get(key string, f func() []byte) []byte {

	if elt := c.t1.Lookup(key); elt != nil {
		c.t1.Remove(key, elt)
		c.t2.PushFront(key)
		return c.data[key]
	}

	if elt := c.t2.Lookup(key); elt != nil {
		c.t2.MoveToFront(elt)
		return c.data[key]
	}

	result := f()
	c.data[key] = result

	if elt := c.b1.Lookup(key); elt != nil {
		c.part = min(c.cap, c.part+max(c.b2.Len()/c.b1.Len(), 1))
		c.replace(key)
		c.b1.Remove(key, elt)
		c.t2.PushFront(key)
		return result
	}

	if elt := c.b2.Lookup(key); elt != nil {
		c.part = max(0, c.part-max(c.b1.Len()/c.b2.Len(), 1))
		c.replace(key)
		c.b2.Remove(key, elt)
		c.t2.PushFront(key)
		return result
	}

	if c.t1.Len()+c.b1.Len() == c.cap {
		if c.t1.Len() < c.cap {
			c.b1.RemoveTail()
			c.replace(key)
		} else {
			pop := c.t1.RemoveTail()
			delete(c.data, pop)
		}
	} else {
		total := c.t1.Len() + c.b1.Len() + c.t2.Len() + c.b2.Len()
		if total >= c.cap {
			if total == (2 * c.cap) {
				c.b2.RemoveTail()
			}
			c.replace(key)
		}
	}

	c.t1.PushFront(key)

	return result
}

func min(x, y int) int {

	if x < y {
		return x
	}

	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}

	return y
}
