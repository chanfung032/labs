package skiplist

import (
	"math/rand"
	"testing"
)

func TestEmpty(t *testing.T) {
	s := New()
	if _, ok := s.Search(1); ok {
		t.Error("Found not existed key")
	}
	if ok := s.Delete(1); ok {
		t.Error("Delete not existed key successed")
	}
}

func TestSeq(t *testing.T) {
	s := New()
	for i := 0; i < 100; i++ {
		s.Insert(i, i)
		if v, ok := s.Search(i); !ok || v != i {
			t.Errorf("Existed key not found: expect %d, get %v, %v", i, v, ok)
		}
	}

	for i := 0; i < 100; i++ {
		if ok := s.Delete(i); !ok {
			t.Errorf("Delete existed key failed")
		}
	}
	for i := 0; i < 100; i++ {
		if _, ok := s.Search(i); ok {
			t.Errorf("Not existed key found: %d", i)
		}
	}
}

func TestRand(t *testing.T) {
	s := New()
	var keys []int
	for i := 0; i < 1000; i++ {
		r := rand.Int()
		keys = append(keys, r)
		s.Insert(r, r)
		if v, ok := s.Search(r); !ok || v != r {
			t.Errorf("Existed key not found: expect %d, get %v, %v", i, v, ok)
		}
	}
	for _, i := range keys {
		if ok := s.Delete(i); !ok {
			t.Errorf("Delete existed key failed")
		}
	}
	for _, i := range keys {
		if _, ok := s.Search(i); ok {
			t.Errorf("Not existed key found: %d", i)
		}
	}
}
