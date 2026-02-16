package gore

import "testing"

// TestMatchSimple tests basic literal matching and dot metacharacter
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

// TestMatchAlternation tests the | operator
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

// TestMatchCharClass tests character classes and ranges
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

// TestExtendedEscapes tests \D, \W, \S and literal escapes
func TestExtendedEscapes(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		// Negated character classes
		{"\\D", "a", true},   // non-digit
		{"\\D", "5", false},  // digit
		{"\\W", "!", true},   // non-word
		{"\\W", "a", false},  // word char
		{"\\S", "a", true},   // non-space
		{"\\S", " ", false},  // space
		{"\\S", "\t", false}, // tab (whitespace)

		// Literal escapes
		{"\\n", "\n", true},
		{"\\t", "\t", true},
		{"\\r", "\r", true},
		{"hello\\nworld", "hello\nworld", true},
		{"tab\\there", "tab\there", true},

		// Escaped metacharacters
		{"\\.", ".", true},
		{"\\.", "a", false},
		{"\\*", "*", true},
		{"\\+", "+", true},
		{"\\?", "?", true},
		{"\\[", "[", true},
		{"\\\\", "\\", true},

		// Combined patterns
		{"\\d+\\s+\\w+", "123 hello", true},
		{"\\D+", "hello", true},
		{"\\W+", "!!!", true},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestCharacterClassEscapes tests escaped sequences within character classes
func TestCharacterClassEscapes(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		// Escapes in character classes
		{"[\\n\\t]", "\n", true},
		{"[\\n\\t]", "\t", true},
		{"[\\n\\t]", "n", false},

		// Escaped metacharacters in classes
		{"[\\[\\]]", "[", true},
		{"[\\[\\]]", "]", true},
		{"[a\\-z]", "-", true},
		{"[a\\-z]", "a", true},
		{"[a\\-z]", "b", false}, // not a range

		// Negated with escapes
		{"[^\\d]", "a", true},
		{"[^\\d]", "5", false},

		// Mixed
		{"[a-z\\d]", "5", true},
		{"[a-z\\d]", "m", true},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tt.pattern, tt.input, got, tt.want)
		}
	}
}
