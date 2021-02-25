package ac

// https://leetcode-cn.com/problems/stream-of-characters/

type StreamChecker struct {
	root    *Node
	current *Node
}

func Constructor(words []string) StreamChecker {
	root := Compile(words)
	return StreamChecker{root, root}
}

func (this *StreamChecker) Query(letter byte) bool {
	node := this.current

	for node.children[letter] == nil && node != this.root {
		node = node.fail
	}
	node = node.children[letter]
	if node == nil {
		this.current = this.root
		return false
	}
	this.current = node

	if node.match {
		return true
	}
	for suffix := node.matchSuffix; suffix != nil; suffix = suffix.matchSuffix {
		return true
	}
	return false
}
