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
	cached map[string][]byte

	c int
	p int

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

func (c *clist) RemoveKey(key string) {
	elt, ok := c.keys[key]
	if !ok {
		panic("removing unavailable key")
	}
	delete(c.keys, key)
	c.l.Remove(elt)
}

func (c *clist) PushFront(key string) {
	elt := c.l.PushFront(key)
	c.keys[key] = elt
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
		cached: make(map[string][]byte),
		c:      size,
		t1:     newClist(),
		t2:     newClist(),
		b1:     newClist(),
		b2:     newClist(),
	}
}

func (self *Cache) replace(key string) {

	var old string
	if (self.t1.Len() > 0 && self.b2.Has(key) && self.t1.Len() == self.p) || (self.t1.Len() > self.p) {
		old = self.t1.RemoveTail()
		self.b1.PushFront(old)
	} else {
		old = self.t2.RemoveTail()
		self.b2.PushFront(old)
	}

	delete(self.cached, old)
}

// Get retrieves a value from the cache. The function f will be called to retrieve the value if it is not present in the cache.
func (self *Cache) Get(key string, f func() []byte) []byte {

	if self.t1.Has(key) {
		self.t1.RemoveKey(key)
		self.t2.PushFront(key)
		return self.cached[key]
	}

	if self.t2.Has(key) {
		self.t2.RemoveKey(key)
		self.t2.PushFront(key)
		return self.cached[key]
	}

	result := f()
	self.cached[key] = result

	if self.b1.Has(key) {
		self.p = min(self.c, self.p+max(self.b2.Len()/self.b1.Len(), 1))
		self.replace(key)
		self.b1.RemoveKey(key)
		self.t2.PushFront(key)
		return result
	}

	if self.b2.Has(key) {
		self.p = max(0, self.p-max(self.b1.Len()/self.b2.Len(), 1))
		self.replace(key)
		self.b2.RemoveKey(key)
		self.t2.PushFront(key)
		return result
	}

	if self.t1.Len()+self.b1.Len() == self.c {
		if self.t1.Len() < self.c {
			self.b1.RemoveTail()
			self.replace(key)
		} else {
			pop := self.t1.RemoveTail()
			delete(self.cached, pop)
		}
	} else {
		total := self.t1.Len() + self.b1.Len() + self.t2.Len() + self.b2.Len()
		if total >= self.c {
			if total == (2 * self.c) {
				self.b2.RemoveTail()
			}
			self.replace(key)
		}
	}

	self.t1.PushFront(key)

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
