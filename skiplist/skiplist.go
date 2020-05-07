package skiplist

import (
	"math/rand"
)

type SkipList struct {
	header *node
	level  int
}

type node struct {
	key     int
	value   interface{}
	forward []*node
}

const (
	p        = 1 / 2.0
	maxLevel = 16
)

func New() *SkipList {
	return &SkipList{
		header: &node{forward: make([]*node, 1)},
	}
}

func randomLevel() int {
	level := 1
	// rand.Float32() 返回一个 [0.0, 1.0) 区间的随机数
	for rand.Float32() < p && level < maxLevel {
		level += 1
	}
	return level
}

func (s *SkipList) Search(searchKey int) (interface{}, bool) {
	x := s.header
	for i := s.level - 1; i >= 0; i-- {
		for x.forward[i] != nil && x.forward[i].key < searchKey {
			x = x.forward[i]
		}
	}
	x = x.forward[0]

	if x != nil && x.key == searchKey {
		return x.value, true
	} else {
		return nil, false
	}
}

func (s *SkipList) Insert(searchKey int, newValue interface{}) {
	var update [maxLevel]*node

	x := s.header
	for i := s.level - 1; i >= 0; i-- {
		for x.forward[i] != nil && x.forward[i].key < searchKey {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]

	if x != nil && x.key == searchKey {
		x.value = newValue
	} else {
		level := randomLevel()

		if level > s.level {
			forward := s.header.forward
			if cap(forward) < level {
				s.header.forward = make([]*node, level)
				copy(s.header.forward, forward)
			}
			for i := s.level; i < level; i++ {
				update[i] = s.header
			}
			s.level = level
		}

		x = &node{searchKey, newValue, make([]*node, level)}
		for i := 0; i < level; i++ {
			x.forward[i] = update[i].forward[i]
			update[i].forward[i] = x
		}
	}
}

func (s *SkipList) Delete(searchKey int) bool {
	var update [maxLevel]*node
	x := s.header
	for i := s.level - 1; i >= 0; i-- {
		for x.forward[i] != nil && x.forward[i].key < searchKey {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]
	if x != nil && x.key == searchKey {
		for i := 0; i < s.level; i++ {
			if update[i].forward[i] != x {
				break
			}
			update[i].forward[i] = x.forward[i]
		}

		for s.level > 0 && s.header.forward[s.level-1] == nil {
			s.level--
		}
		return true
	} else {
		return false
	}
}
