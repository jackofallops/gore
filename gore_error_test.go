package gore

import "testing"

// TestInvalidPatterns tests that invalid regex patterns produce errors
func TestInvalidPatterns(t *testing.T) {
	invalidPatterns := []struct {
		pattern string
		desc    string
		skip    bool // true if validation not yet implemented
	}{
		{"(", "unclosed group", false},
		{")", "unmatched closing paren", false},
		{"[", "unclosed character class", false},
		{"[z-a]", "invalid range", false},                                    // FIXED
		{"(?P<>abc)", "empty capture name", false},                           // FIXED
		{"(?P<123>abc)", "invalid capture name (starts with digit)", false},  // FIXED
		{"(?P<name>a)(?P<name>b)", "duplicate capture name", false},          // FIXED
		{"*", "quantifier without target", false},                            // FIXED
		{"+", "quantifier without target", false},                            // FIXED
		{"?", "quantifier without target", false},                            // FIXED
		{"{3}", "quantifier without target", false},                          // FIXED
		{"(?", "incomplete group", false},
		{"(?P", "incomplete named group", false},
		{"\\", "trailing backslash", false},
		{"[\\", "unclosed escape in class", false},
		{"a{", "unclosed quantifier", false},
		{"a{3,2}", "invalid range (min > max)", false},                       // FIXED
		{"(?P<name)", "incomplete named group", false},
	}

	for _, tt := range invalidPatterns {
		if tt.skip {
			t.Logf("SKIP: Compile(%q) should fail (%s) - validation not yet implemented",
				tt.pattern, tt.desc)
			continue
		}
		_, err := Compile(tt.pattern)
		if err == nil {
			t.Errorf("Compile(%q) should fail (%s), but succeeded",
				tt.pattern, tt.desc)
		}
	}
}

// TestValidEdgeCasePatterns tests valid patterns that might seem unusual
func TestValidEdgeCasePatterns(t *testing.T) {
	validPatterns := []struct {
		pattern string
		input   string
		want    bool
	}{
		{"", "", true},           // empty pattern
		{"", "a", true},          // empty pattern matches anywhere
		{"(?:)", "", true},       // empty non-capturing group
		{"()", "", true},         // empty capturing group
		{"a{0}", "", true},       // zero repetitions
		{"a{0,0}", "", true},     // zero to zero
		{"a{0}b", "b", true},     // zero repetitions before b
		{"x{1,1}", "x", true},    // single repetition range
		{"(?i:a)", "A", true},    // case insensitive non-capturing group
		{"(?i)", "", true},       // flag-only group
	}

	for _, tt := range validPatterns {
		re, err := Compile(tt.pattern)
		if err != nil {
			t.Errorf("Compile(%q) should succeed, but failed: %v",
				tt.pattern, err)
			continue
		}
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("Pattern %q on input %q: got %v, want %v",
				tt.pattern, tt.input, got, tt.want)
		}
	}
}
