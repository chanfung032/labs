package main

// https://swtch.com/~rsc/regexp/regexp1.html
// https://swtch.com/~rsc/regexp/nfa.c.txt

import "fmt"

// 将正则表达式的中缀形式转换为后缀形式
func Re2Post(re string) string {
	type pframe struct {
		nalt  int
		natom int
	}
	var paren []pframe
	var buf []byte
	var (
		nalt  int // | 分支计数器
		natom int // | 操作数计数器
	)

	for _, c := range []byte(re) {
		switch c {
		case '(':
			if natom > 1 {
				natom--
				buf = append(buf, '.')
			}
			paren = append(paren, pframe{nalt, natom})
			nalt = 0
			natom = 0
		case ')':
			if len(paren) == 0 {
				return ""
			}
			if natom == 0 {
				return ""
			}
			for natom--; natom > 0; natom-- {
				buf = append(buf, '.')
			}
			for ; nalt > 0; nalt-- {
				buf = append(buf, '|')
			}
			p := paren[len(paren)-1]
			paren = paren[:len(paren)-1]
			nalt = p.nalt
			natom = p.natom
			natom++
		case '|':
			if natom == 0 {
				return ""
			}
			for natom--; natom > 0; natom-- {
				buf = append(buf, '.')
			}
			nalt++
		case '*':
			fallthrough
		case '+':
			fallthrough
		case '?':
			if natom == 0 {
				return ""
			}
			buf = append(buf, c)
		default:
			// 两个操作数（atom）之间加一个显式的连接操作符
			if natom > 1 {
				natom--
				buf = append(buf, '.')
			}
			buf = append(buf, c)
			natom++
		}
	}

	if len(paren) != 0 {
		return ""
	}
	for natom--; natom > 0; natom-- {
		buf = append(buf, '.')
	}
	for ; nalt > 0; nalt-- {
		buf = append(buf, '|')
	}
	return string(buf)
}

const (
	Match = iota
	Split
)

type State struct {
	c    byte
	out  *State
	out1 *State
}

var matchstate = State{c: Match}

func Post2Nfa(postfix string) *State {
	type Frag struct {
		start *State
		out   []**State
	}
	var stack []Frag
	push := func(f Frag) {
		stack = append(stack, f)
	}
	pop := func() (f Frag) {
		f = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return
	}

	patch := func(out []**State, s *State) {
		for _, p := range out {
			*p = s
		}
	}

	for _, c := range []byte(postfix) {
		switch c {
		default:
			s := &State{c: c}
			push(Frag{s, []**State{&s.out}})
		case '.':
			e2 := pop()
			e1 := pop()
			patch(e1.out, e2.start)
			push(Frag{e1.start, e2.out})
		case '|':
			e2 := pop()
			e1 := pop()
			s := &State{c: Split, out: e1.start, out1: e2.start}
			push(Frag{s, append(e1.out, e2.out...)})
		case '?':
			e := pop()
			s := &State{c: Split, out: e.start}
			push(Frag{s, append(e.out, &s.out1)})
		case '*':
			e := pop()
			s := &State{c: Split, out: e.start}
			patch(e.out, s)
			push(Frag{s, []**State{&s.out1}})
		case '+':
			e := pop()
			s := &State{c: Split, out: e.start}
			patch(e.out, s)
			push(Frag{e.start, []**State{&s.out1}})
		}
	}

	e := pop()
	if len(stack) != 0 {
		return nil
	}
	patch(e.out, &matchstate)
	return e.start
}

func addstate(states map[*State]bool, s *State) {
	if s == nil || states[s] {
		return
	}
	if s.c == Split {
		addstate(states, s.out)
		addstate(states, s.out1)
	}
	states[s] = true
}

func step(states map[*State]bool, c byte) map[*State]bool {
	nstates := make(map[*State]bool)
	for s, _ := range states {
		if s.c == c {
			addstate(nstates, s.out)
		}
	}
	return nstates
}

func IsMatch(start *State, s string) bool {
	states := map[*State]bool{}
	addstate(states, start)
	for _, c := range []byte(s) {
		states = step(states, c)
	}
	fmt.Println("states:", states)
	return states[&matchstate]
}

// 使用graphviz dot输出states状态图
func PrintStates(s *State) {
	states := map[*State]bool{}

	q := []*State{s}

	fmt.Println("digraph D {\nrankdir=LR\nnode [shape=circle]")
	for len(q) != 0 {
		s := q[0]
		q = q[1:]
		if states[s] {
			continue
		}
		states[s] = true
		if s.c == Split {
			fmt.Printf("s%p [label=\"\" color=\"red\"]\n", s)
		} else if s.c == Match {
			fmt.Printf("s%p [label=\"M\" color=\"red\" shape=doublecircle]\n", s)
		} else {
			fmt.Printf("s%p [label=\"%c\"]\n", s, s.c)
		}
		if s.out != nil {
			fmt.Printf("s%p -> s%p\n", s, s.out)
			q = append(q, s.out)
		}
		if s.out1 != nil {
			fmt.Printf("s%p -> s%p\n", s, s.out1)
			q = append(q, s.out1)
		}
	}
	fmt.Println("}")
}
