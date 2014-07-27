// Package arc implements the Adaptive Replacement Cache
/*

This code is a straight-forward translation of the Python implementation at http://code.activestate.com/recipes/576532-adaptive-replacement-cache-in-python/

*/
package arc

/*
things to clean up:
    s/self/cache/
    b1Keys map[string]*Element
*/

import (
	"container/list"
)

type Cache struct {
	cached map[string][]byte

	c int
	p int

	t1 *list.List
	t2 *list.List
	b1 *list.List
	b2 *list.List
}

func New(size int) *Cache {
	return &Cache{
		cached: make(map[string][]byte),
		c:      size,
		t1:     list.New(),
		t2:     list.New(),
		b1:     list.New(),
		b2:     list.New(),
	}
}

// to make first draft python translation easier
// TODO(dgryski): replace these with map[string]*Element to make this O(1) instead of O(n)

func listIn(l *list.List, key string) bool {

	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == key {
			return true
		}
	}

	return false
}

func listRemove(l *list.List, key string) {

	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == key {
			l.Remove(e)
			return
		}
	}

	panic("key not found in remove")
}

func (self *Cache) replace(key string) {

	var old string
	if (self.t1.Len() > 0 && listIn(self.b2, key) && self.t1.Len() == self.p) || (self.t1.Len() > self.p) {
		old = self.t1.Remove(self.t1.Back()).(string)
		self.b1.PushFront(old)
	} else {
		old = self.t2.Remove(self.t2.Back()).(string)
		self.b2.PushFront(old)
	}

	delete(self.cached, old)
}

func (self *Cache) Get(key string, f func() []byte) []byte {

	if listIn(self.t1, key) {
		listRemove(self.t1, key)
		self.t2.PushFront(key)
		return self.cached[key]
	}

	if listIn(self.t2, key) {
		listRemove(self.t2, key)
		self.t2.PushFront(key)
		return self.cached[key]
	}

	result := f()
	self.cached[key] = result

	if listIn(self.b1, key) {
		self.p = min(self.c, self.p+max(self.b2.Len()/self.b1.Len(), 1))
		self.replace(key)
		listRemove(self.b1, key)
		self.t2.PushFront(key)
		return result
	}

	if listIn(self.b2, key) {
		self.p = max(0, self.p-max(self.b1.Len()/self.b2.Len(), 1))
		self.replace(key)
		listRemove(self.b2, key)
		self.t2.PushFront(key)
		return result
	}

	if self.t1.Len()+self.b1.Len() == self.c {
		if self.t1.Len() < self.c {
			self.b1.Remove(self.b1.Back())
			self.replace(key)
		} else {
			pop := self.t1.Remove(self.t1.Back()).(string)
			delete(self.cached, pop)
		}
	} else {
		total := self.t1.Len() + self.b1.Len() + self.t2.Len() + self.b2.Len()
		if total >= self.c {
			if total == (2 * self.c) {
				self.b2.Remove(self.b2.Back())
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
