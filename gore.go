package gore

import (
	"fmt"
	"io"
)

type Regexp struct {
	expr        string
	prog        *Prog
	subexpNames []string
}

func Compile(expr string) (*Regexp, error) {
	parser := NewParser(expr)
	node, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	compiler := NewCompiler()
	prog, err := compiler.Compile(node, parser.captures)
	if err != nil {
		return nil, err
	}

	// Build subexp names from parser
	names := make([]string, parser.captures+1)
	for name, idx := range parser.names {
		if idx < len(names) {
			names[idx] = name
		}
	}

	return &Regexp{
		expr:        expr,
		prog:        prog,
		subexpNames: names,
	}, nil
}

func MustCompile(expr string) *Regexp {
	re, err := Compile(expr)
	if err != nil {
		panic(fmt.Sprintf("gore: Compile(%q): %v", expr, err))
	}
	return re
}

func (re *Regexp) SubexpNames() []string {
	return re.subexpNames
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

func (re *Regexp) FindStringSubmatch(s string) []string {
	input := NewStringInput(s)
	vm := NewVM(re.prog, input)

	// Unanchored search through input (including EOF for empty matches)
	inputLen := input.Len()
	pos := 0
	for pos <= inputLen {
		// Use prefix search to skip impossible positions
		if re.prog.Prefix != "" && pos < inputLen {
			prefixPos := input.Index(re, pos)
			if prefixPos == -1 {
				return nil // No prefix found
			}
			pos = prefixPos
		}

		matched, caps := vm.Run(pos)
		if matched {
			// Build result from captures
			result := make([]string, len(re.subexpNames))
			for i := 0; i < len(result); i++ {
				start, end := -1, -1
				if 2*i < len(caps) {
					start = caps[2*i]
				}
				if 2*i+1 < len(caps) {
					end = caps[2*i+1]
				}
				if start >= 0 && end >= 0 && end >= start {
					result[i] = s[start:end]
				}
			}
			return result
		}

		_, w := input.Step(pos)
		if w == 0 {
			break
		}
		pos += w
	}
	return nil
}

func (re *Regexp) match(input Input) bool {
	vm := NewVM(re.prog, input)
	inputLen := input.Len()

	pos := 0
	for pos <= inputLen {
		// Use prefix search to skip impossible positions
		if re.prog.Prefix != "" && pos < inputLen {
			prefixPos := input.Index(re, pos)
			if prefixPos == -1 {
				return false // No prefix found anywhere
			}
			pos = prefixPos
		}

		matched, _ := vm.Run(pos)
		if matched {
			return true
		}
		_, w := input.Step(pos)
		if w == 0 {
			break
		}
		pos += w
	}
	return false
}
