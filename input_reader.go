package gore

import (
	"io"
	"io/ioutil"
	"unicode/utf8"
)

// ReaderInput implements Input for an io.Reader.
// Currently reads all input into memory to support backtracking.
type ReaderInput struct {
	data []byte
}

func NewReaderInput(r io.Reader) (*ReaderInput, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &ReaderInput{data: b}, nil
}

func (s *ReaderInput) Step(pos int) (rune, int) {
	if pos >= len(s.data) {
		return 0, 0
	}
	r, w := utf8.DecodeRune(s.data[pos:])
	return r, w
}

func (s *ReaderInput) Context(pos int) (rune, int) {
	if pos <= 0 {
		return -1, 0
	}
	if pos > len(s.data) {
		pos = len(s.data)
	}
	r, w := utf8.DecodeLastRune(s.data[:pos])
	return r, w
}

func (s *ReaderInput) Index(re *Regexp, pos int) int {
	return -1
}
