package gore

import (
	"strings"
	"testing"
)

func BenchmarkLiteral(b *testing.B) {
	re := MustCompile("abc")
	input := "xabcy"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

func BenchmarkLookahead(b *testing.B) {
	re := MustCompile(`q(?=u)`)
	input := "quit"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

func BenchmarkLookbehind(b *testing.B) {
	re := MustCompile(`(?<=foo)bar`)
	input := "foobar"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkLookbehindLong checks performance scaling of the naive O(N) lookbehind.
// Since it scans from the start for every position, a long prefix should hurt.
func BenchmarkLookbehindLongPrefix(b *testing.B) {
	re := MustCompile(`(?<=foo)bar`)
	payload := strings.Repeat("x", 1000)
	input := payload + "foobar"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkPathological tests a nested quantifier case that triggers exponential backtracking.
// Pattern: (a+)+b against aaaaa...a
func BenchmarkPathological(b *testing.B) {
	// A modest N to avoid stalling the test suite too long,
	// but enough to show the drop in performance compared to linear engines.
	re := MustCompile(`(a+)+b`)
	input := "aaaaaaaaaaaaaaaaaaaa" // 20 'a's
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkNamedCaptures benchmarks the performance of named capture groups.
func BenchmarkNamedCaptures(b *testing.B) {
	re := MustCompile(`(?P<first>\w+)\s+(?P<last>\w+)`)
	input := "John Doe"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.FindStringSubmatch(input)
	}
}
