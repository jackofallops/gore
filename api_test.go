package gore

import (
	"reflect"
	"testing"
)

// TestFindString tests basic find functionality
func TestFindString(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    string
	}{
		{"world", "hello world", "world"},
		{"\\d+", "abc123def", "123"},
		{"[a-z]+", "123abc456", "abc"},
		{"notfound", "hello world", ""},
		{"^start", "start here", "start"},
		{"end$", "the end", "end"},
		{"a*", "", ""}, // Empty match at start
		{"(?<=foo)bar", "foobar", "bar"},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.FindString(tt.input)
		if got != tt.want {
			t.Errorf("FindString(%q, %q) = %q; want %q", tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestFindStringIndex tests finding match indices
func TestFindStringIndex(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    []int
	}{
		{"world", "hello world", []int{6, 11}},
		{"\\d+", "abc123def", []int{3, 6}},
		{"[a-z]+", "123abc456", []int{3, 6}},
		{"notfound", "hello world", nil},
		{"^start", "start here", []int{0, 5}},
		{"end$", "the end", []int{4, 7}},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.FindStringIndex(tt.input)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FindStringIndex(%q, %q) = %v; want %v", tt.pattern, tt.input, got, tt.want)
		}
	}
}

// TestFindAllStringSubmatch tests finding multiple matches with submatches
func TestFindAllStringSubmatch(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		n       int
		want    [][]string
	}{
		// Find all words
		{"\\w+", "hello world foo", -1, [][]string{{"hello"}, {"world"}, {"foo"}}},

		// Find first 2 digits
		{"\\d", "a1b2c3", 2, [][]string{{"1"}, {"2"}}},

		// With captures
		{"(\\w+)=(\\d+)", "a=1 b=2 c=3", -1, [][]string{
			{"a=1", "a", "1"},
			{"b=2", "b", "2"},
			{"c=3", "c", "3"},
		}},

		// n=0 returns nil
		{"\\w+", "hello", 0, nil},

		// No matches
		{"\\d+", "abc", -1, nil},

		// Empty matches (lookahead)
		{"(?=\\w)", "ab", 2, [][]string{{""}, {""}}},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.FindAllStringSubmatch(tt.input, tt.n)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FindAllStringSubmatch(%q, %q, %d) = %v; want %v",
				tt.pattern, tt.input, tt.n, got, tt.want)
		}
	}
}

// TestFindAllStringIndex tests finding all match indices
func TestFindAllStringIndex(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		n       int
		want    [][]int
	}{
		// Find all words
		{"\\w+", "hello world", -1, [][]int{{0, 5}, {6, 11}}},

		// Find first 2
		{"\\d", "a1b2c3", 2, [][]int{{1, 2}, {3, 4}}},

		// n=0 returns nil
		{"\\w+", "hello", 0, nil},

		// No matches
		{"\\d+", "abc", -1, nil},

		// Adjacent matches
		{"a", "aaa", -1, [][]int{{0, 1}, {1, 2}, {2, 3}}},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.FindAllStringIndex(tt.input, tt.n)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FindAllStringIndex(%q, %q, %d) = %v; want %v",
				tt.pattern, tt.input, tt.n, got, tt.want)
		}
	}
}

// TestSplit tests string splitting
func TestSplit(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		n       int
		want    []string
	}{
		// Split on comma
		{",", "a,b,c", -1, []string{"a", "b", "c"}},

		// Split with limit
		{",", "a,b,c,d", 2, []string{"a", "b,c,d"}},

		// Split on whitespace
		{"\\s+", "hello  world\tfoo", -1, []string{"hello", "world", "foo"}},

		// n=0 returns nil
		{",", "a,b,c", 0, nil},

		// No matches returns original
		{",", "abc", -1, []string{"abc"}},

		// Split at start
		{",", ",a,b", -1, []string{"", "a", "b"}},

		// Split at end
		{",", "a,b,", -1, []string{"a", "b", ""}},

		// Multiple consecutive separators
		{",", "a,,b", -1, []string{"a", "", "b"}},
	}

	for _, tt := range tests {
		re := MustCompile(tt.pattern)
		got := re.Split(tt.input, tt.n)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Split(%q, %q, %d) = %v; want %v",
				tt.pattern, tt.input, tt.n, got, tt.want)
		}
	}
}

// TestFindAllWithBoundedQuantifiers tests FindAll with new {n,m} syntax
func TestFindAllWithBoundedQuantifiers(t *testing.T) {
	re := MustCompile("\\d{2,3}")
	got := re.FindAllStringSubmatch("1 12 123 1234", -1)
	want := [][]string{{"12"}, {"123"}, {"123"}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindAllStringSubmatch with {n,m} = %v; want %v", got, want)
	}
}

// TestFindAllWithWordBoundaries tests FindAll with \b
func TestFindAllWithWordBoundaries(t *testing.T) {
	re := MustCompile("\\bcat\\b")
	got := re.FindAllStringSubmatch("cat scat cats caterpillar", -1)
	want := [][]string{{"cat"}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindAllStringSubmatch with \\b = %v; want %v", got, want)
	}
}

// TestSplitWithComplexPattern tests Split with captures and lookarounds
func TestSplitWithComplexPattern(t *testing.T) {
	// Split on word boundaries
	re := MustCompile("\\s+")
	got := re.Split("The quick brown fox", -1)
	want := []string{"The", "quick", "brown", "fox"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Split on whitespace = %v; want %v", got, want)
	}
}
