package gore

import (
	"unicode/utf8"
)

// Input abstracts the source of text to be matched.
// It allows the regex engine to work transparently with strings, byte slices, and io.Readers.
type Input interface {
	// Step returns the rune at the given position and its width in bytes.
	// If the position is at or beyond the end of the input, it returns (0, 0).
	// It basically acts like utf8.DecodeRune.
	Step(pos int) (rune, int)

	// Context returns the rune before the given position, to support boundary checks like \b and ^.
	// If pos is 0, it should return (-1, 0) or handling for start-of-text.
	Context(pos int) (rune, int)

	// Index returns the byte index of the given string/pattern in the input starting at pos.
	// Used for optimizations (prefix search). Returns -1 if not found.
	Index(re *Regexp, pos int) int
}

// StringInput implements Input for a string.
type StringInput struct {
	str string
}

func NewStringInput(s string) *StringInput {
	return &StringInput{str: s}
}

func (s *StringInput) Step(pos int) (rune, int) {
	if pos >= len(s.str) {
		return 0, 0
	}
	r, w := utf8.DecodeRuneInString(s.str[pos:])
	return r, w
}

func (s *StringInput) Context(pos int) (rune, int) {
	if pos <= 0 {
		return -1, 0
	}
	if pos > len(s.str) {
		pos = len(s.str)
	}
	r, w := utf8.DecodeLastRuneInString(s.str[:pos])
	return r, w
}

func (s *StringInput) Index(re *Regexp, pos int) int {
	return -1
}
