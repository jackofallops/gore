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

func TestFindStringSubmatch(t *testing.T) {
	tests := []struct {
		pattern  string
		input    string
		expected []string
	}{
		{
			`(\w+)\s+(\w+)`,
			"John Doe",
			[]string{"John Doe", "John", "Doe"},
		},
		{
			`(?P<first>\w+)\s+(?P<last>\w+)`,
			"Jane Smith",
			[]string{"Jane Smith", "Jane", "Smith"},
		},
		{
			`a(b*)c`,
			"abbbc",
			[]string{"abbbc", "bbb"},
		},
		{
			`a(b*)c`,
			"ac",
			[]string{"ac", ""},
		},
	}
	for _, tc := range tests {
		re := MustCompile(tc.pattern)
		got := re.FindStringSubmatch(tc.input)
		// Check lengths
		if len(got) != len(tc.expected) {
			t.Errorf("FindStringSubmatch(%q, %q) length = %d; want %d. Got: %v", tc.pattern, tc.input, len(got), len(tc.expected), got)
			continue
		}
		// Check content
		for i, s := range got {
			if s != tc.expected[i] {
				t.Errorf("FindStringSubmatch(%q, %q)[%d] = %q; want %q", tc.pattern, tc.input, i, s, tc.expected[i])
			}
		}
	}
}

func TestSubexpNames(t *testing.T) {
	pattern := `(?P<first>\w+)\s+(\w+)\s+(?P<last>\w+)`
	re := MustCompile(pattern)
	names := re.SubexpNames()
	// capturing groups are:
	// 1: first (\w+)
	// 2: (\w+) (unnamed)
	// 3: last (\w+)
	// Index 0 is implicit whole match (empty name usually in Go stdlib).
	expected := []string{"", "first", "", "last"}
	if len(names) != len(expected) {
		t.Fatalf("SubexpNames length = %d; want %d", len(names), len(expected))
	}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("SubexpNames[%d] = %q; want %q", i, name, expected[i])
		}
	}
}
