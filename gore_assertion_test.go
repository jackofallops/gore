package gore

import (
	"strings"
	"testing"
)

// TestLookahead tests positive and negative lookahead assertions
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

// TestLookbehind tests positive and negative lookbehind assertions
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

// TestWordBoundaries tests \b and \B
func TestWordBoundaries(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		// \b - word boundary
		{"\\bword\\b", "word", true},
		{"\\bword\\b", "word.", true},
		{"\\bword\\b", " word ", true},
		{"\\bword\\b", "sword", false},
		{"\\bword\\b", "words", false},
		{"\\bword\\b", "wording", false},

		// Start boundary
		{"\\bcat", "cat", true},
		{"\\bcat", "category", true},
		{"\\bcat", "scat", false},

		// End boundary
		{"cat\\b", "cat", true},
		{"cat\\b", "scat", true},
		{"cat\\b", "cats", false},

		// \B - NOT a word boundary
		{"\\Bcat", "cat", false},
		{"\\Bcat", "scat", true},
		{"cat\\B", "cats", true},
		{"cat\\B", "cat", false},

		// Complex patterns
		{"\\b\\d+\\b", "123", true},
		{"\\b\\d+\\b", "abc123def", false},
		{"\\b\\w+\\b", "hello world", true},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestStringAnchors tests \A, \Z, \z anchors (distinct from ^ and $)
func TestStringAnchors(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		// \A matches start of string only
		{"\\Astart", "start", true},
		{"\\Astart", "\nstart", false},

		// \Z matches end before final newline
		{"end\\Z", "end\n", true},
		{"end\\Z", "end", true},
		{"end\\Z", "end\nmore", false},

		// \z matches absolute end
		{"end\\z", "end", true},
		{"end\\z", "end\n", false},
		{"end\\Z", "end\n", true}, // comparison
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestMatchReader tests matching from io.Reader
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
