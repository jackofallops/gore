package gore

import (
	"strings"
	"testing"
)

func TestMatchSimple(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		match   bool
	}{
		{"abc", "abc", true},
		{"abc", "xabcy", true},
		{"abc", "ab", false},
		{"a.c", "abc", true},
		{"a.c", "axc", true},
		{"a.c", "ac", false}, // dot needs char
	}

	for _, tc := range tests {
		re := MustCompile(tc.pattern)
		if got := re.MatchString(tc.input); got != tc.match {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tc.pattern, tc.input, got, tc.match)
		}
	}
}

func TestMatchAlternation(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		match   bool
	}{
		{"a|b", "a", true},
		{"a|b", "b", true},
		{"a|b", "c", false},
		{"foo|bar", "foo", true},
		{"foo|bar", "bar", true},
		{"foo|bar", "baz", false},
	}
	for _, tc := range tests {
		re := MustCompile(tc.pattern)
		if got := re.MatchString(tc.input); got != tc.match {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tc.pattern, tc.input, got, tc.match)
		}
	}
}

func TestMatchQuantifier(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		match   bool
	}{
		{"a*", "", true},
		{"a*", "aaaa", true},
		{"a+", "a", true},
		{"a+", "", false},
		{"a?", "", true},
		{"a?", "a", true},
		{"a?", "aa", true}, // matches 'a' subset
	}
	for _, tc := range tests {
		re := MustCompile(tc.pattern)
		if got := re.MatchString(tc.input); got != tc.match {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tc.pattern, tc.input, got, tc.match)
		}
	}
}

func TestMatchCharClass(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		match   bool
	}{
		{"[a-z]", "a", true},
		{"[a-z]", "A", false},
		{"[a-z]", "z", true},
		{"[^a-z]", "A", true},
		{"[^a-z]", "a", false},
	}
	for _, tc := range tests {
		re := MustCompile(tc.pattern)
		if got := re.MatchString(tc.input); got != tc.match {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tc.pattern, tc.input, got, tc.match)
		}
	}
}

func TestLookahead(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		match   bool
	}{
		{"a(?=b)", "ab", true},
		{"a(?=b)", "ac", false},
		{"a(?!b)", "ac", true},
		{"a(?!b)", "ab", false},
		{"q(?=u)", "quit", true},
		{"q(?!u)", "quote", false},
	}
	for _, tc := range tests {
		re := MustCompile(tc.pattern)
		if got := re.MatchString(tc.input); got != tc.match {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tc.pattern, tc.input, got, tc.match)
		}
	}
}

func TestLookbehind(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		match   bool
	}{
		{"(?<=a)b", "ab", true},
		{"(?<=a)b", "cb", false},
		{"(?<!a)b", "cb", true},
		{"(?<!a)b", "ab", false},
		{"(?<=foo)bar", "foobar", true},
	}
	for _, tc := range tests {
		re := MustCompile(tc.pattern)
		if got := re.MatchString(tc.input); got != tc.match {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tc.pattern, tc.input, got, tc.match)
		}
	}
}

func TestMatchReader(t *testing.T) {
	re := MustCompile("hello world")
	r := strings.NewReader("hello world")
	matched, err := re.MatchReader(r)
	if err != nil {
		t.Fatalf("MatchReader error: %v", err)
	}
	if !matched {
		t.Error("MatchReader failed to match")
	}
}
