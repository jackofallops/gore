package gore

import (
	"fmt"
	"strings"
	"testing"
)

// TestStressLongInput tests performance and correctness with very long inputs
func TestStressLongInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Very long input
	re := MustCompile("needle")
	haystack := strings.Repeat("x", 100000) + "needle"

	if !re.MatchString(haystack) {
		t.Error("Should find needle in very long string")
	}

	// Test with FindString
	got := re.FindString(haystack)
	if got != "needle" {
		t.Errorf("FindString in long input: got %q; want %q", got, "needle")
	}
}

// TestStressComplexPattern tests patterns with many alternations
func TestStressComplexPattern(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Pattern with many alternations
	alternatives := make([]string, 100)
	for i := range alternatives {
		alternatives[i] = fmt.Sprintf("word%d", i)
	}
	pattern := strings.Join(alternatives, "|")

	re := MustCompile(pattern)
	if !re.MatchString("word50") {
		t.Error("Should match in large alternation")
	}

	// Test matching first and last alternatives
	if !re.MatchString("word0") {
		t.Error("Should match first alternative")
	}
	if !re.MatchString("word99") {
		t.Error("Should match last alternative")
	}
}

// TestStressNestedGroups tests deeply nested capture groups
func TestStressNestedGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Deeply nested groups
	depth := 20
	pattern := strings.Repeat("(", depth) + "a" + strings.Repeat(")", depth)
	re, err := Compile(pattern)
	if err != nil {
		t.Fatalf("Should compile nested groups: %v", err)
	}

	if !re.MatchString("a") {
		t.Error("Nested groups should match")
	}

	// Check capture groups
	matches := re.FindStringSubmatch("a")
	expectedLen := depth + 1 // one for full match, one for each group
	if len(matches) != expectedLen {
		t.Errorf("Expected %d capture groups, got %d", expectedLen, len(matches))
	}
}

// TestStressRepeatedQuantifiers tests patterns with large quantifier values
func TestStressRepeatedQuantifiers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Large bounded quantifier
	re := MustCompile("a{1000}")
	input := strings.Repeat("a", 1000)

	if !re.MatchString(input) {
		t.Error("Should match exactly 1000 'a's")
	}

	// Should not match fewer
	if re.MatchString(strings.Repeat("a", 999)) {
		t.Error("Should not match 999 'a's")
	}
}

// TestStressLongCharacterClass tests character classes with many ranges
func TestStressLongCharacterClass(t *testing.T) {
	// Large character class
	re := MustCompile("[a-zA-Z0-9_!@#$%^&*()\\-+={}\\[\\]:;\"'<>,.?/\\\\|`~]")

	testChars := "abcXYZ123!@#"
	for _, ch := range testChars {
		if !re.MatchString(string(ch)) {
			t.Errorf("Should match character %q", ch)
		}
	}
}

// TestStressMultipleBackreferences tests multiple backreferences in a pattern
func TestStressMultipleBackreferences(t *testing.T) {
	// Pattern with multiple backreferences
	re := MustCompile(`(.)(.)(.)(.)\4\3\2\1`)
	if !re.MatchString("abcddcba") {
		t.Error("Should match palindrome-like pattern")
	}

	if re.MatchString("abcdabcd") {
		t.Error("Should not match non-palindrome")
	}
}
