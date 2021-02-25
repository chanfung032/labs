package ac

import "testing"

func compare(a, b []Match) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func search(text string, words ...string) []Match {
	trie := Compile(words)
	return trie.Search(text)
}

// Text and indexes
// a b x a b c d e f a b
// 0 1 2 3 4 5 6 7 8 9 0
func verify(t *testing.T, words []string, expected []Match) {
	text := "abxabcdefab"
	res := search(text, words...)

	if !compare(res, expected) {
		t.Errorf(
			"Search(%s, %s) returned %v expected %v",
			text, words, res, expected,
		)
	}
}

func TestSearchSinglePosition(t *testing.T) {
	verify(t, []string{"c"}, []Match{Match{"c", 5, 6}})
}

func TestSearchSingleWord(t *testing.T) {
	verify(t, []string{"abc"}, []Match{Match{"abc", 3, 6}})
}

func TestSearchManyWords(t *testing.T) {
	verify(t,
		[]string{"aac", "ac", "abc", "bca"},
		[]Match{Match{"abc", 3, 6}},
	)
}

func TestSearchManyMatches(t *testing.T) {
	verify(t,
		[]string{"ab"},
		[]Match{Match{"ab", 0, 2}, Match{"ab", 3, 5}, Match{"ab", 9, 11}},
	)

	verify(t,
		[]string{"abc", "fab"},
		[]Match{Match{"abc", 3, 6}, Match{"fab", 8, 11}},
	)

	verify(t,
		[]string{"ab", "cde"},
		[]Match{
			Match{"ab", 0, 2}, Match{"ab", 3, 5},
			Match{"cde", 5, 8}, Match{"ab", 9, 11}},
	)
}

func TestPrefixWord(t *testing.T) {
	verify(t,
		[]string{"ab", "abcd"},
		[]Match{
			Match{"ab", 0, 2}, Match{"ab", 3, 5},
			Match{"abcd", 3, 7}, Match{"ab", 9, 11}},
	)

	verify(t,
		[]string{"abcd", "ab"},
		[]Match{
			Match{"ab", 0, 2}, Match{"ab", 3, 5},
			Match{"abcd", 3, 7}, Match{"ab", 9, 11}},
	)
}

func TestPrefixTwice(t *testing.T) {
	verify(t,
		[]string{"ab", "abc", "abcd"},
		[]Match{
			Match{"ab", 0, 2}, Match{"ab", 3, 5}, Match{"abc", 3, 6},
			Match{"abcd", 3, 7}, Match{"ab", 9, 11}},
	)

	verify(t,
		[]string{"abcd", "abc", "ab"},
		[]Match{
			Match{"ab", 0, 2}, Match{"ab", 3, 5}, Match{"abc", 3, 6},
			Match{"abcd", 3, 7}, Match{"ab", 9, 11}},
	)
}

func TestSuffixWord(t *testing.T) {
	verify(t,
		[]string{"ab", "b"},
		[]Match{
			Match{"ab", 0, 2}, Match{"b", 1, 2}, Match{"ab", 3, 5},
			Match{"b", 4, 5}, Match{"ab", 9, 11}, Match{"b", 10, 11}},
	)

	verify(t,
		[]string{"abx", "bx"},
		[]Match{Match{"abx", 0, 3}, Match{"bx", 1, 3}},
	)
}

func TestSuffixTwice(t *testing.T) {
	verify(t,
		[]string{"abx", "bx", "x"},
		[]Match{Match{"abx", 0, 3}, Match{"bx", 1, 3}, Match{"x", 2, 3}},
	)
}

func TestMoveToSuffix(t *testing.T) {
	verify(t,
		[]string{"abxb", "bxa"},
		[]Match{Match{"bxa", 1, 4}},
	)
}

func TestMoveToSuffixTwice(t *testing.T) {
	verify(t,
		[]string{"abxb", "bxb", "xa"},
		[]Match{Match{"xa", 2, 4}},
	)
}

func TestWikipediaExample(t *testing.T) {
	// From https://en.wikipedia.org/wiki/Aho%E2%80%93Corasick_algorithm

	res := search(
		"abccab",
		"a", "ab", "bab", "bc", "bca", "c", "caa")

	expected := []Match{
		Match{"a", 0, 1}, Match{"ab", 0, 2}, Match{"bc", 1, 3},
		Match{"c", 2, 3}, Match{"c", 3, 4},
		Match{"a", 4, 5}, Match{"ab", 4, 6},
	}

	if !compare(res, expected) {
		t.Errorf("Got result %v expected %v", res, expected)
	}
}
