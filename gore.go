package gore

import (
	"fmt"
	"io"
)

type Regexp struct {
	expr string
	prog *Prog
}

func Compile(expr string) (*Regexp, error) {
	parser := NewParser(expr)
	node, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	compiler := NewCompiler()
	prog, err := compiler.Compile(node)
	if err != nil {
		return nil, err
	}

	return &Regexp{
		expr: expr,
		prog: prog,
	}, nil
}

func MustCompile(expr string) *Regexp {
	re, err := Compile(expr)
	if err != nil {
		panic(fmt.Sprintf("gore: Compile(%q): %v", expr, err))
	}
	return re
}

func (re *Regexp) MatchString(s string) bool {
	input := NewStringInput(s)
	return re.match(input)
}

func (re *Regexp) MatchReader(r io.Reader) (bool, error) {
	input, err := NewReaderInput(r)
	if err != nil {
		return false, err
	}
	return re.match(input), nil
}

func (re *Regexp) match(input Input) bool {
	vm := NewVM(re.prog, input)

	// Unanchored search
	// TODO: Use Input length if known?
	// For ReaderInput using ReadAll, we know length implicitly via Step returning EOF.

	// We need a way to know end of input to stop loop.
	// Input interface doesn't expose Len(). Steps until EOF.

	pos := 0
	for {
		matched, _ := vm.Run(pos)
		if matched {
			return true
		}

		// Advance
		_, w := input.Step(pos)
		if w == 0 {
			break
		}
		pos += w
	}
	return false
}
