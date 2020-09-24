package ac

import "fmt"

const debug = false

type Node struct {
	children    map[byte]*Node
	match       bool
	fail        *Node
	matchSuffix *Node
	depth       int
}

func Compile(words []string) *Node {
	root := buildTrie(words)
	fixFailAndMatchSuffix(root)
	return root
}

func newNode(depth int) *Node {
	return &Node{children: make(map[byte]*Node), depth: depth}
}

func buildTrie(words []string) *Node {
	root := newNode(0)

	for _, word := range words {
		node := root

		for i := 0; i < len(word); i++ {
			ch := word[i]
			child, exist := node.children[ch]
			if !exist {
				child = newNode(i + 1)
				node.children[ch] = child
			}
			node = child
		}

		node.match = true
	}
	return root
}

func fixFailAndMatchSuffix(root *Node) {
	var queue []*Node

	for _, child := range root.children {
		queue = append(queue, child)
		child.fail = root
	}
	printNode(root)

	for len(queue) != 0 {
		var node *Node
		node, queue = queue[0], queue[1:]

		for ch, child := range node.children {
			queue = append(queue, child)

			fail := node.fail
			for fail.children[ch] == nil && fail != root {
				fail = fail.fail
			}
			child.fail = fail.children[ch]
			if child.fail == nil {
				child.fail = root
			}
		}

		for next := node.fail; next != nil; next = next.fail {
			if next.match {
				node.matchSuffix = next
				break
			}
		}

		printNode(node)
	}
}

type Match struct {
	word  string
	start int
	end   int
}

func newMatch(text string, begin, end int) Match {
	return Match{text[begin:end], begin, end}
}

func (root *Node) Search(text string) []Match {
	var m []Match

	node := root
	for i := 0; i < len(text); i++ {
		ch := text[i]

		for node.children[ch] == nil && node != root {
			node = node.fail
		}
		node = node.children[ch]
		if node == nil {
			node = root
			continue
		}

		if node.match {
			m = append(m, newMatch(text, i-node.depth+1, i+1))
		}
		for suffix := node.matchSuffix; suffix != nil; suffix = suffix.matchSuffix {
			m = append(m, newMatch(text, i-suffix.depth+1, i+1))
		}
	}

	return m
}

func printNode(node *Node) {
	if !debug {
		return
	}
	for ch, child := range node.children {
		fmt.Printf("\"%p\" -> \"%p\" [label=\"%c\"]\n", node, child, ch)
	}
	if node.match {
		fmt.Printf("\"%p\" [label=\"\", style=filled, fillcolor=yellow]\n", node)
	} else {
		fmt.Printf("\"%p\" [label=\"\"]\n", node)
	}
	if node.fail != nil {
		fmt.Printf("\"%p\" -> \"%p\" [style=dotted]\n", node, node.fail)
	}
	if node.matchSuffix != nil {
		fmt.Printf("\"%p\" -> \"%p\" [color=blue]\n", node, node.matchSuffix)
	}
}
