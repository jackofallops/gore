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

// TestBoundedQuantifiers tests {n}, {n,m}, and {n,} syntax
func TestBoundedQuantifiers(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		// {n} - exactly n times
		{"a{3}", "aaa", true},
		{"a{3}", "aa", false},
		{"a{3}", "aaaa", true}, // matches first 3
		{"^a{3}$", "aaaa", false},
		{"^a{3}$", "aaa", true},

		// {n,m} - between n and m times
		{"a{2,4}", "a", false},
		{"a{2,4}", "aa", true},
		{"a{2,4}", "aaa", true},
		{"a{2,4}", "aaaa", true},
		{"a{2,4}", "aaaaa", true}, // matches first 4
		{"^a{2,4}$", "aaaaa", false},

		// {n,} - n or more times
		{"a{3,}", "aa", false},
		{"a{3,}", "aaa", true},
		{"a{3,}", "aaaa", true},
		{"a{3,}", "aaaaaaaa", true},

		// Non-greedy variants
		{"a{2,4}?", "aaaa", true}, // should match, still

		// Complex patterns
		{"[0-9]{3}-[0-9]{4}", "123-4567", true},
		{"[0-9]{3}-[0-9]{4}", "12-4567", false},
		{"\\d{2,3}", "12", true},
		{"\\d{2,3}", "123", true},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v", tt.pattern, tt.input, got, tt.want)
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

func TestComplexRegexPatterns(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		subject    string
		wantMatch  bool
		wantGroups []string // The full match is index 0
	}{
		// --- 1. POSITIVE & NEGATIVE LOOKAHEADS ---
		// Validates passwords: 8-16 chars, 1+ upper, 1+ digit, NO whitespace
		{
			name:      "Password Validator - Valid",
			pattern:   `^(?=.*[A-Z])(?=.*\d)(?![^a-zA-Z\d]*\s)\S{8,16}$`,
			subject:   "Secure7890!",
			wantMatch: true,
		},
		{
			name:      "Password Validator - Invalid (No Digit)",
			pattern:   `^(?=.*[A-Z])(?=.*\d)(?![^a-zA-Z\d]*\s)\S{8,16}$`,
			subject:   "OnlyLetters!",
			wantMatch: false,
		},

		// --- 2. POSITIVE LOOKBEHIND ---
		// Extract domain names only if preceded by https://
		{
			name:       "Secure Domain Extraction",
			pattern:    `(?<=https:\/\/)([\w.-]+)\.(com|org|net)`,
			subject:    "Visit https://api.github.com for details",
			wantMatch:  true,
			wantGroups: []string{"api.github.com", "api.github", "com"},
		},

		// --- 3. NEGATIVE LOOKBEHIND & LOOKAHEAD ---
		// Matches dollar amounts that are NOT negative (debt) and have exactly 2 decimals
		{
			name:       "Strict Currency Filter",
			pattern:    `(?<!-)\$\s?(\d+(?:\.\d{2})(?=\s|$))`,
			subject:    "The balance is $1250.00 today",
			wantMatch:  true,
			wantGroups: []string{"$1250.00", "1250.00"},
		},
		{
			name:      "Strict Currency Filter - Skip Negative",
			pattern:   `(?<!-)\$\s?(\d+(?:\.\d{2})(?=\s|$))`,
			subject:   "Debt: -$50.00",
			wantMatch: false,
		},

		// --- 4. BACKREFERENCES (Stress Test) ---
		// Matches HTML/XML tags ensuring the closing tag matches the opening tag
		{
			name:       "Tag Matching Backreference",
			pattern:    `<([a-z1-6]+)>.*?</\1>`,
			subject:    "<h1>Welcome to GORE</h1>",
			wantMatch:  true,
			wantGroups: []string{"<h1>Welcome to GORE</h1>", "h1"},
		},
		{
			name:      "Tag Mismatch",
			pattern:   `<([a-z1-6]+)>.*?<\/\1>`,
			subject:   "<div>Mismatch</span>",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile the pattern
			re, err := Compile(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to compile pattern %q: %v", tt.pattern, err)
			}

			// Perform match
			match := re.MatchString(tt.subject)
			if match != tt.wantMatch {
				t.Errorf("MatchString(%q) = %v; want %v", tt.subject, match, tt.wantMatch)
			}

			// Check capture groups if a match was expected
			if tt.wantMatch && len(tt.wantGroups) > 0 {
				gotGroups := re.FindStringSubmatch(tt.subject)
				if len(gotGroups) != len(tt.wantGroups) {
					t.Errorf("Expected %d capture groups, got %d", len(tt.wantGroups), len(gotGroups))
				}
				for i, want := range tt.wantGroups {
					if i < len(gotGroups) && gotGroups[i] != want {
						t.Errorf("Group %d = %q; want %q", i, gotGroups[i], want)
					}
				}
			}
		})
	}
}