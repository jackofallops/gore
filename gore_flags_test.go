package gore

import "testing"

// TestCaseInsensitive tests the (?i) flag
func TestCaseInsensitive(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		// Basic ASCII
		{"(?i)abc", "ABC", true},
		{"(?i)abc", "abc", true},
		{"(?i)ABC", "abc", true},
		{"(?i)aBc", "AbC", true},

		// Mixed case
		{"abc", "ABC", false},

		// Scoped flags
		{"(?i)abc(?-i)def", "ABCdef", true},
		{"(?i)abc(?-i)def", "ABCDEF", false},
		{"(?i)abc(?-i)DEF", "ABCdef", false},

		// Character classes
		{"(?i)[a-z]", "A", true},
		{"(?i)[A-Z]", "a", true},
		{"(?i)[a-z]+", "HELLO", true},
		{"(?i)[^a-z]", "A", false}, // negated matches 'A' because 'a' matches 'A'
		{"(?i)[^0-9]", "A", true},

		// Unicode
		{"(?i)k", "\u212A", true}, // Kelvin sign K matches k
		{"(?i)\u212A", "k", true},
		{"(?i)s", "\u017F", true}, // long s matches s

		// Combinations
		{"(?i)a+", "AAA", true},
		{"(?i)(abc)+", "ABCabcABC", true},

		// Escaped characters
		{"(?i)\\w", "A", true},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestCaseInsensitiveReplace tests replacement with case-insensitivity
func TestCaseInsensitiveReplace(t *testing.T) {
	re := MustCompile("(?i)apple")
	got := re.ReplaceAllString("Apple apple APPLE", "orange")
	want := "orange orange orange"

	if got != want {
		t.Errorf("ReplaceAllString = %q; want %q", got, want)
	}
}

// TestCaseInsensitiveGroup tests localized flags
func TestCaseInsensitiveGroup(t *testing.T) {
	// (?i:...) non-capturing group with flag
	re := MustCompile("(?i:abc)def")

	if !re.MatchString("ABCdef") {
		t.Error("Should match ABCdef")
	}
	if re.MatchString("ABCDEF") {
		t.Error("Should not match ABCDEF")
	}
}

// TestMultilineMode tests the (?m) flag for multiline matching
func TestMultilineMode(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		// Default: ^ and $ match string boundaries
		{"^line", "first\nline", false},

		// Multiline: ^ and $ match line boundaries
		{"(?m)^line", "first\nline", true},
		{"(?m)end$", "end\nmore", true},
		{"end$", "end\nmore", false},

		// Multiple lines
		{"(?m)^\\w+", "one\ntwo\nthree", true}, // should match "one"
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v",
				tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestDotallMode tests the (?s) flag for dotall matching
func TestDotallMode(t *testing.T) {
	tests := []struct{
		pattern string
		input   string
		want    bool
	}{
		// Default: . doesn't match newline
		{"a.b", "a\nb", false},

		// Dotall: . matches newline
		{"(?s)a.b", "a\nb", true},
		{"(?s).*", "line1\nline2", true},

		// Combined modes
		{"(?ms)^.*$", "line1\nline2", true},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v",
				tt.pattern, tt.input, got, tt.want)
		}
	}
}
