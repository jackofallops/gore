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

// BenchmarkCharClass benchmarks basic character class matching.
func BenchmarkCharClass(b *testing.B) {
	re := MustCompile("[a-zA-Z0-9_]+")
	input := "hello_world_123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkNegatedCharClass benchmarks negated character class matching.
func BenchmarkNegatedCharClass(b *testing.B) {
	re := MustCompile("[^0-9]+")
	input := "abcdefghijklmnop"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkBoundedQuantifier benchmarks bounded quantifier patterns like {n,m}.
func BenchmarkBoundedQuantifier(b *testing.B) {
	re := MustCompile("[0-9]{3}-[0-9]{4}")
	input := "123-4567"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkAlternation benchmarks alternation (|) performance with multiple branches.
func BenchmarkAlternation(b *testing.B) {
	re := MustCompile("foo|bar|baz")
	input := "baz"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkNegativeLookahead benchmarks negative lookahead assertions.
func BenchmarkNegativeLookahead(b *testing.B) {
	re := MustCompile(`a(?!b)`)
	input := "ac"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkNegativeLookbehind benchmarks negative lookbehind assertions.
func BenchmarkNegativeLookbehind(b *testing.B) {
	re := MustCompile(`(?<!a)b`)
	input := "cb"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkWordBoundary benchmarks word boundary (\b) matching.
func BenchmarkWordBoundary(b *testing.B) {
	re := MustCompile(`\bword\b`)
	input := "find word in text"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkBackreferences benchmarks backreference performance in patterns.
func BenchmarkBackreferences(b *testing.B) {
	re := MustCompile(`<([a-z]+)>.*?</\1>`)
	input := "<div>content</div>"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkQuantifierStar benchmarks star (*) quantifier with long input.
func BenchmarkQuantifierStar(b *testing.B) {
	re := MustCompile("a*b")
	input := strings.Repeat("a", 100) + "b"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

// BenchmarkQuantifierPlus benchmarks plus (+) quantifier with long input.
func BenchmarkQuantifierPlus(b *testing.B) {
	re := MustCompile("a+b")
	input := strings.Repeat("a", 100) + "b"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}
