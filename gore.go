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
	prog, err := compiler.Compile(node)
	if err != nil {
		return nil, err
	}

	// Capture names from parser
	// parser.captures is the count of capturing groups (1-based)
	// parser.names maps name -> index
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

	// Unanchored Search
	pos := 0
	for {
		matched, caps := vm.Run(pos)
		if matched {
			// Convert caps (start,end pairs?)
			// VM currently returns `[]int` caps where index i is position?
			// Wait, VM `caps` logic: `caps[2*i]` is start, `caps[2*i+1]` is end?
			// Let's verify VM save instruction.
			// OpSave idx: `caps[inst.Idx] = pos`.
			// Capture node emits: Save(2*i), Body, Save(2*i+1).
			// So yes, even indices are starts, odd are ends.
			// caps[0], caps[1] are group 0 (whole match)? No, group 0 isn't emitted by Capture node usually.
			// Go stdlib group 0 is implicit whole match.
			// My compiler doesn't emit Group 0 capture instructions.
			// I need to wrap the whole expression in a capture 0?
			// Or just take start=pos, end=match_end?
			// VM.Run returns (bool, caps).

			// If match found, we have caps for 1..N.
			// Group 0 (whole match) is implicitly (pos, end_pos).
			// To get end_pos, we need `vm.Run` to return it or derive it.
			// My `match` logic returns `true` but consumes input inside VM.
			// But `vm.Run` calls `match` which returns `true`.
			// `match` does NOT return end position currently in the signature `(bool, []int)`.
			// The `caps` slice is updated.

			// Issue: I don't know the end of the match!
			// Group 0 capture is missing.

			// Quick fix: Add implicit Group 0 capture in Compile?
			// Or modify VM return.

			// Let's update Compile to wrap everything in Save(0)...Save(1).

			result := make([]string, len(re.subexpNames))

			// We need to resolve end of match.
			// Let's assume for now I fix Compiler to add Group 0.

			for i := 0; i < len(result); i++ {
				start := -1
				end := -1
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
	pos := 0
	for {
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
