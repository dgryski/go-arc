package arc

import (
	"container/list"
	"encoding/binary"
	"testing"
)

func TestARC(t *testing.T) {

	tst := []uint32{
		// doctest from the python code
		// range(20) + range(11,15) + range(20) + range(11,40) + [39, 38, 37, 36, 35, 34, 33, 32, 16, 17, 11, 41]]
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		11, 12, 13, 14,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 39,
		38, 37, 36, 35, 34, 33, 32, 16, 17, 11, 41,
	}

	cache := New(10)

	for _, v := range tst {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], v)
		cache.Get(string(b[:]), func() interface{} { return b[:] })
	}

	checkList(t, "t1", cache.t1.l, []byte{41})
	checkList(t, "t2", cache.t2.l, []byte{11, 17, 16, 32, 33, 34, 35, 36, 37})
	checkList(t, "b1", cache.b1.l, []byte{31, 30})
	checkList(t, "b2", cache.b2.l, []byte{38, 39, 19, 18, 15, 14, 13, 12})

	if cache.part != 5 {
		t.Errorf("bad p: got=%v want=5", cache.part)
	}
}

func checkList(t *testing.T, name string, l *list.List, expected []byte) {

	idx := 0

	for e := l.Front(); e != nil; e = e.Next() {
		b := []byte(e.Value.(string))
		if b[0] != expected[idx] {
			t.Errorf("list %s failed idx %d: got=%d want=%d\n", name, idx, b[0], expected[idx])
		}
		idx++
	}
}
