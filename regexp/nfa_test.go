package main

import (
	"fmt"
	"testing"
)

func TestRe2post(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"a", "a"},
		{"ab", "ab."},
		{"abcdef", "ab.c.d.e.f."},
		{"a|b", "ab|"},
		{"a|b|c|e|f", "abcef||||"},
		{"a*", "a*"},
		{"a?", "a?"},
		{"a+", "a+"},
		{"a*|a?|a+", "a*a?a+||"},
		{"(a*|a?|a+)?|(a*|a?|a+)*|(a*|a?|a+)+", "a*a?a+||?a*a?a+||*a*a?a+||+||"},
		{"(((a?)b+)c*)abc", "a?b+.c*.a.b.c."},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out := Re2Post(tt.in)
			if out != tt.out {
				t.Errorf("want %v, got %v", tt.out, out)
			}
		})
	}
}

func TestIsMatch(t *testing.T) {
	tests := []struct {
		p string
		s []string
		n []string
	}{
		{"abc", []string{"abc"}, []string{"ab", "a", "abcd"}},
		{"a*", []string{"", "a", "aaaa"}, []string{"ab", "aab"}},
		{"a+", []string{"a", "aaa"}, []string{""}},
		{"a?", []string{"a", ""}, []string{"aa"}},
		{"a|b|c|d", []string{"a", "b", "c", "d"}, []string{"e"}},
		{"(a|b)*", []string{"", "a", "b", "aa", "ab", "ba"}, []string{"ac"}},
		{"a?a?a?aaa", []string{"aaa", "aaaa", "aaaaaa"}, []string{"aa", "aaaaaaa"}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt), func(t *testing.T) {
			start := Post2Nfa(Re2Post(tt.p))
			for _, s := range tt.s {
				out := IsMatch(start, s)
				if out != true {
					t.Errorf("p:%v, s:%v, want true, got %v", tt.p, s, out)
				}
			}
			for _, s := range tt.n {
				out := IsMatch(start, s)
				if out != false {
					t.Errorf("p:%v, s:%v, want false, got %v", tt.p, s, out)
				}
			}
		})
	}
}
