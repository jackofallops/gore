package gore

import "testing"

// TestReplaceAllString tests basic string replacement
func TestReplaceAllString(t *testing.T) {
	tests := []struct {
		pattern string
		src     string
		repl    string
		want    string
	}{
		// Simple replacement
		{"world", "hello world", "Go", "hello Go"},

		// With captures
		{`(\w+)@(\w+)`, "user@example", "$2.$1", "example.user"},

		// Multiple matches
		{`\d+`, "a1b2c3", "X", "aXbXcX"},

		// Named captures
		{`(?P<user>\w+)@(?P<domain>\w+)`, "john@example", "$domain/$user", "example/john"},

		// $$ escaping
		{`\d+`, "price: 100", "$$$$", "price: $$"},

		// ${name} syntax
		{`(?P<num>\d+)`, "x=123", "${num}0", "x=1230"},

		// No match
		{`\d+`, "abc", "X", "abc"},

		// Empty replacement
		{`\s+`, "a  b  c", "", "abc"},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.ReplaceAllString(tt.src, tt.repl)
		if got != tt.want {
			t.Errorf("ReplaceAllString(%q, %q, %q) = %q; want %q",
				tt.pattern, tt.src, tt.repl, got, tt.want)
		}
	}
}

// TestReplaceAllLiteralString tests literal replacement (no expansion)
func TestReplaceAllLiteralString(t *testing.T) {
	re := MustCompile(`\d+`)
	got := re.ReplaceAllLiteralString("a1b2c3", "$1")
	want := "a$1b$1c$1"

	if got != want {
		t.Errorf("ReplaceAllLiteralString = %q; want %q", got, want)
	}
}

// TestReplaceAllStringFunc tests function-based replacement
func TestReplaceAllStringFunc(t *testing.T) {
	re := MustCompile(`\d+`)

	got := re.ReplaceAllStringFunc("a1b22c333", func(s string) string {
		return "[" + s + "]"
	})
	want := "a[1]b[22]c[333]"

	if got != want {
		t.Errorf("ReplaceAllStringFunc = %q; want %q", got, want)
	}
}

// TestReplaceComplexPattern tests replacement with complex regex
func TestReplaceComplexPattern(t *testing.T) {
	// Test with lookaround
	re := MustCompile(`\b\w+\b`)
	got := re.ReplaceAllString("hello world", "[$0]")
	want := "[hello] [world]"

	if got != want {
		t.Errorf("Replace with \\b = %q; want %q", got, want)
	}
}

// TestReplaceEdgeCases tests edge cases in template expansion
func TestReplaceEdgeCases(t *testing.T) {
	tests := []struct {
		pattern string
		src     string
		repl    string
		want    string
	}{
		// Trailing $
		{`\w+`, "hello", "x$", "x$"},

		// Invalid capture group
		{`(\w+)`, "hello", "$2", ""},

		// Unclosed ${
		{`\w+`, "hello", "${", "${"},

		// Empty submatch (matches at multiple positions)
		{`(\w*)`, "a", "[$1]", "[a][]"},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.ReplaceAllString(tt.src, tt.repl)
		if got != tt.want {
			t.Errorf("ReplaceAllString(%q, %q, %q) = %q; want %q",
				tt.pattern, tt.src, tt.repl, got, tt.want)
		}
	}
}

// TestReplaceAll tests byte slice variant
func TestReplaceAll(t *testing.T) {
	re := MustCompile(`\d+`)
	got := string(re.ReplaceAll([]byte("a1b2c3"), []byte("X")))
	want := "aXbXcX"

	if got != want {
		t.Errorf("ReplaceAll = %q; want %q", got, want)
	}
}
