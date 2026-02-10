package gore

import (
	"reflect"
	"testing"
)

// TestFind tests byte slice Find methods
func TestFind(t *testing.T) {
	re := MustCompile(`\d+`)

	got := re.Find([]byte("abc123def"))
	want := []byte("123")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Find = %v; want %v", got, want)
	}

	// No match
	got = re.Find([]byte("abc"))
	if got != nil {
		t.Errorf("Find (no match) = %v; want nil", got)
	}
}

// TestFindIndex tests byte slice FindIndex
func TestFindIndex(t *testing.T) {
	re := MustCompile(`\d+`)

	got := re.FindIndex([]byte("abc123def"))
	want := []int{3, 6}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindIndex = %v; want %v", got, want)
	}
}

// TestFindSubmatch tests byte slice FindSubmatch
func TestFindSubmatch(t *testing.T) {
	re := MustCompile(`(\w+)@(\w+)`)

	got := re.FindSubmatch([]byte("user@example.com"))
	want := [][]byte{
		[]byte("user@example"),
		[]byte("user"),
		[]byte("example"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindSubmatch = %v; want %v", got, want)
	}
}

// TestFindAll tests byte slice FindAll
func TestFindAll(t *testing.T) {
	re := MustCompile(`\d+`)

	got := re.FindAll([]byte("a1b22c333"), -1)
	want := [][]byte{
		[]byte("1"),
		[]byte("22"),
		[]byte("333"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindAll = %v; want %v", got, want)
	}

	// With limit
	got = re.FindAll([]byte("a1b22c333"), 2)
	want = [][]byte{
		[]byte("1"),
		[]byte("22"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindAll (limit) = %v; want %v", got, want)
	}
}

// TestFindAllIndex tests byte slice FindAllIndex
func TestFindAllIndex(t *testing.T) {
	re := MustCompile(`\w+`)

	got := re.FindAllIndex([]byte("hello world"), -1)
	want := [][]int{{0, 5}, {6, 11}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindAllIndex = %v; want %v", got, want)
	}
}

// TestFindAllSubmatch tests byte slice FindAllSubmatch
func TestFindAllSubmatch(t *testing.T) {
	re := MustCompile(`(\w+)=(\d+)`)

	got := re.FindAllSubmatch([]byte("a=1 b=2"), -1)
	want := [][][]byte{
		{[]byte("a=1"), []byte("a"), []byte("1")},
		{[]byte("b=2"), []byte("b"), []byte("2")},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindAllSubmatch = %v; want %v", got, want)
	}
}

// TestMatch tests byte slice Match
func TestMatch(t *testing.T) {
	re := MustCompile(`\d+`)

	if !re.Match([]byte("abc123")) {
		t.Error("Match should be true")
	}

	if re.Match([]byte("abc")) {
		t.Error("Match should be false")
	}
}
