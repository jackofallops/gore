package gore

import "testing"

// TestMatchQuantifier tests basic quantifiers *, +, ?
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

// TestNonGreedyQuantifiers tests non-greedy quantifier behavior
func TestNonGreedyQuantifiers(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    string
	}{
		// Non-greedy star
		{"a.*?b", "axxxbxxxb", "axxxb"},
		{"a.*b", "axxxbxxxb", "axxxbxxxb"}, // greedy comparison

		// Non-greedy plus
		{".+?", "abc", "a"}, // matches minimal
		{".+", "abc", "abc"}, // greedy comparison

		// Non-greedy question mark
		{"a??b", "ab", "ab"},

		// Non-greedy bounded
		{"a{2,4}?", "aaaaa", "aa"},   // should match minimum
		{"a{2,4}", "aaaaa", "aaaa"},  // greedy comparison

		// Practical example: HTML tags
		{"<.*?>", "<a>text</a>", "<a>"},
		{"<.*>", "<a>text</a>", "<a>text</a>"},

		// Non-greedy with longer match needed
		{"a.*?c", "abc", "abc"},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.FindString(tt.input)
		if got != tt.want {
			t.Errorf("FindString(%q, %q) = %q; want %q", tt.pattern, tt.input, got, tt.want)
		}
	}
}
