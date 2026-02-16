package gore

import "testing"

// TestFindStringSubmatch tests basic capture group functionality
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

// TestSubexpNames tests named capture group names
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

// TestNonCapturingGroups tests (?:...) syntax
func TestNonCapturingGroups(t *testing.T) {
	// (?:...) should not create capture groups
	re := MustCompile(`(?:foo|bar)(\d+)`)
	matches := re.FindStringSubmatch("foo123")

	// Should have 2 elements: full match and one capture
	if len(matches) != 2 {
		t.Errorf("Expected 2 groups, got %d: %v", len(matches), matches)
	}
	if matches[0] != "foo123" {
		t.Errorf("Full match = %q; want %q", matches[0], "foo123")
	}
	if matches[1] != "123" {
		t.Errorf("Capture 1 = %q; want %q", matches[1], "123")
	}

	// Nested non-capturing groups
	re2 := MustCompile(`(?:a(?:b|c))(d)`)
	matches2 := re2.FindStringSubmatch("abd")
	if len(matches2) != 2 {
		t.Errorf("Nested: expected 2 groups, got %d", len(matches2))
	}
}

// TestNestedCaptureGroups tests nested capturing groups
func TestNestedCaptureGroups(t *testing.T) {
	tests := []struct {
		pattern  string
		input    string
		expected []string
	}{
		// Nested groups
		{
			`((a)(b))`,
			"ab",
			[]string{"ab", "ab", "a", "b"},
		},
		// Deeply nested
		{
			`(a(b(c)))`,
			"abc",
			[]string{"abc", "abc", "bc", "c"},
		},
		// Mixed nesting
		{
			`(a(b)c)(d(e))`,
			"abcde",
			[]string{"abcde", "abc", "b", "de", "e"},
		},
		// Nested with quantifiers
		{
			`((a)+b)+`,
			"aabaaab",
			[]string{"aabaaab", "aaab", "a"},
		},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.FindStringSubmatch(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("Pattern %q: got %d groups, want %d\nGot: %v\nWant: %v",
				tt.pattern, len(got), len(tt.expected), got, tt.expected)
			continue
		}
		for i, s := range got {
			if s != tt.expected[i] {
				t.Errorf("Pattern %q, group %d = %q; want %q",
					tt.pattern, i, s, tt.expected[i])
			}
		}
	}
}

// TestOptionalGroupsAndBackrefs tests optional groups with backreferences
func TestOptionalGroupsAndBackrefs(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
		skip    bool   // true if behavior not yet correct
		reason  string // reason for skip
	}{
		// Optional capturing group
		{"(a)?b", "b", true, false, ""},
		{"(a)?b", "ab", true, false, ""},

		// Backreference to optional group
		// FIXED: Empty backreferences now correctly fail to match
		{"(a)?(b)\\1", "ba", false, false, ""},
		{"(a)?(b)\\1", "aba", true, false, ""},

		// Multiple backreferences
		{"(.)(.)(.)\\3\\2\\1", "abccba", true, false, ""},
		{"(.)(.)(.)\\3\\2\\1", "abcdef", false, false, ""},

		// Backreference in alternation
		{"(a)\\1|b", "aa", true, false, ""},
		{"(a)\\1|b", "b", true, false, ""},
		{"(a)\\1|b", "a", false, false, ""},
	}

	for _, tt := range tests {
		if tt.skip {
			t.Logf("SKIP: MatchString(%q, %q) - %s", tt.pattern, tt.input, tt.reason)
			continue
		}
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q) = %v; want %v",
				tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestComplexRegexPatterns tests advanced combinations of features
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
