package gore

import "testing"

// TestEmptyMatchesAndZeroWidth tests edge cases with empty and zero-width matches
func TestEmptyMatchesAndZeroWidth(t *testing.T) {
	// Empty pattern
	re := MustCompile("")
	if !re.MatchString("anything") {
		t.Error("Empty pattern should match")
	}

	// Empty group captures
	re2 := MustCompile("a(b*)c")
	matches := re2.FindStringSubmatch("ac")
	if len(matches) != 2 || matches[1] != "" {
		t.Errorf("Empty group: got %v; want [\"ac\", \"\"]", matches)
	}

	// Zero-width assertions don't consume
	re3 := MustCompile("(?=a)a")
	got := re3.FindString("a")
	if got != "a" {
		t.Errorf("Zero-width lookahead consumed: got %q", got)
	}

	// Multiple zero-width matches
	re4 := MustCompile("\\b")
	matches4 := re4.FindAllStringIndex("hello world", -1)
	// Should find 4 boundaries: |hello| |world|
	if len(matches4) != 4 {
		t.Errorf("Word boundaries: got %d matches; want 4", len(matches4))
	}

	// Empty alternation branches
	re5 := MustCompile("a||b")
	if !re5.MatchString("") {
		t.Error("Empty alternation branch should match empty")
	}
}

// TestEmptyStringMatching tests various patterns against empty strings
func TestEmptyStringMatching(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
		skip    bool   // true if behavior not yet correct
		reason  string // reason for skip
	}{
		{"", true, false, ""},
		{"a?", true, false, ""},
		{"a*", true, false, ""},
		{"a+", false, false, ""},
		{"()", true, false, ""},
		{"(?:)", true, false, ""},
		{"^$", true, false, ""},
		{"\\b\\b", false, false, ""}, // CORRECT: word boundaries require word chars (verified with PCRE2/Perl/Python/Go)
		{"(?=a)", false, false, ""},  // lookahead fails on empty
		{"(?!a)", true, false, ""},   // negative lookahead succeeds
	}

	for _, tt := range tests {
		if tt.skip {
			t.Logf("SKIP: Pattern %q on empty string - %s", tt.pattern, tt.reason)
			continue
		}
		re := MustCompile(tt.pattern)
		got := re.MatchString("")
		if got != tt.want {
			t.Errorf("Pattern %q on empty string: got %v; want %v",
				tt.pattern, got, tt.want)
		}
	}
}

// TestConsecutiveWordBoundaries tests consecutive \b behavior matches PCRE2
func TestConsecutiveWordBoundaries(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		// Empty string: no word characters, therefore no word boundaries
		{`\b`, "", false},
		{`\b\b`, "", false},
		{`\b\b\b\b`, "", false},

		// Single word char: has boundaries at start and end (positions 0 and 1)
		{`\b`, "a", true},
		{`\b\b`, "a", true}, // matches at both boundaries
		{`\b\b\b\b`, "a", true},

		// Two word chars: boundaries at positions 0 and 2
		{`\b`, "ab", true},
		{`\b\b`, "ab", true},

		// Word boundary only exists at transitions
		{`\b`, "a b", true}, // matches at any boundary
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("Pattern %q on %q: got %v; want %v",
				tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestZeroWidthAssertionPositions tests that zero-width assertions match at correct positions
func TestZeroWidthAssertionPositions(t *testing.T) {
	// Lookahead at start
	re := MustCompile("^(?=hello)")
	idx := re.FindStringIndex("hello world")
	if idx == nil || idx[0] != 0 || idx[1] != 0 {
		t.Errorf("Lookahead at start: got %v; want [0 0]", idx)
	}

	// Lookbehind at end
	re2 := MustCompile("(?<=world)$")
	idx2 := re2.FindStringIndex("hello world")
	if idx2 == nil || idx2[0] != 11 || idx2[1] != 11 {
		t.Errorf("Lookbehind at end: got %v; want [11 11]", idx2)
	}

	// Word boundaries don't consume
	re3 := MustCompile("\\b")
	match := re3.FindString("abc")
	if match != "" {
		t.Errorf("Word boundary should return empty string, got %q", match)
	}
}
