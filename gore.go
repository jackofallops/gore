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

// NumSubexp returns the number of parenthesized subexpressions in this Regexp.
func (re *Regexp) NumSubexp() int {
	return len(re.subexpNames) - 1
}

// SubexpNames returns the names of the parenthesized subexpressions
// in this Regexp. The first element is the full match (unnamed).
func (re *Regexp) SubexpNames() []string {
	return re.subexpNames
}

// SubexpIndex returns the index of the first subexpression with the given name,
// or -1 if there is no subexpression with that name.
func (re *Regexp) SubexpIndex(name string) int {
	for i, n := range re.subexpNames {
		if n == name {
			return i
		}
	}
	return -1
}

// String returns the source text used to compile the regular expression.
func (re *Regexp) String() string {
	return re.expr
}

// LiteralPrefix returns a literal string that must begin any match
// of the regular expression re. It returns the boolean true if the
// literal string comprises the entire regular expression.
func (re *Regexp) LiteralPrefix() (prefix string, complete bool) {
	if re.prog.Prefix == "" {
		return "", false
	}
	// Check if entire pattern is just this literal
	// For now, we return false for complete since we don't track this
	// This optimization could be added later
	return re.prog.Prefix, false
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

// FindString returns the leftmost match of the regular expression in s.
// Returns empty string if no match found.
func (re *Regexp) FindString(s string) string {
	match := re.FindStringIndex(s)
	if match == nil {
		return ""
	}
	return s[match[0]:match[1]]
}

// FindStringIndex returns a two-element slice of integers defining the location
// of the leftmost match in s. Returns nil if no match found.
func (re *Regexp) FindStringIndex(s string) []int {
	input := NewStringInput(s)
	vm := NewVM(re.prog, input)

	pos := 0
	inputLen := input.Len()
	for pos <= inputLen {
		// Use prefix search if available
		if re.prog.Prefix != "" && pos < inputLen {
			prefixPos := input.Index(re, pos)
			if prefixPos == -1 {
				return nil
			}
			pos = prefixPos
		}

		matched, caps := vm.Run(pos)
		if matched && len(caps) >= 2 {
			return []int{caps[0], caps[1]} // Return [start, end] of whole match
		}

		_, w := input.Step(pos)
		if w == 0 {
			break
		}
		pos += w
	}
	return nil
}

// FindAllStringSubmatch returns a slice of all successive matches of the expression,
// as defined by FindStringSubmatch. n < 0 means return all matches.
func (re *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
	if n == 0 {
		return nil
	}

	var results [][]string
	input := NewStringInput(s)
	inputLen := input.Len()
	pos := 0

	for (n < 0 || len(results) < n) && pos <= inputLen {
		vm := NewVM(re.prog, input)

		// Prefix optimization
		if re.prog.Prefix != "" && pos < inputLen {
			prefixPos := input.Index(re, pos)
			if prefixPos == -1 {
				break
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
			results = append(results, result)

			// Advance past this match (handle zero-width matches)
			matchEnd := caps[1]
			if matchEnd == pos {
				// Zero-width match, advance by one rune
				_, w := input.Step(pos)
				if w == 0 {
					break
				}
				pos += w
			} else {
				pos = matchEnd
			}
		} else {
			_, w := input.Step(pos)
			if w == 0 {
				break
			}
			pos += w
		}
	}

	return results
}

// FindAllStringIndex returns a slice of all successive matches of the expression,
// as two-element slices of integers. n < 0 means return all matches.
func (re *Regexp) FindAllStringIndex(s string, n int) [][]int {
	if n == 0 {
		return nil
	}

	var results [][]int
	input := NewStringInput(s)
	inputLen := input.Len()
	pos := 0

	for (n < 0 || len(results) < n) && pos <= inputLen {
		vm := NewVM(re.prog, input)

		// Prefix optimization
		if re.prog.Prefix != "" && pos < inputLen {
			prefixPos := input.Index(re, pos)
			if prefixPos == -1 {
				break
			}
			pos = prefixPos
		}

		matched, caps := vm.Run(pos)
		if matched && len(caps) >= 2 {
			results = append(results, []int{caps[0], caps[1]})

			// Advance past this match (handle zero-width matches)
			matchEnd := caps[1]
			if matchEnd == pos {
				// Zero-width match, advance by one rune
				_, w := input.Step(pos)
				if w == 0 {
					break
				}
				pos += w
			} else {
				pos = matchEnd
			}
		} else {
			_, w := input.Step(pos)
			if w == 0 {
				break
			}
			pos += w
		}
	}

	return results
}

// Split slices s into substrings separated by the expression and returns a slice of
// the substrings between those expression matches. n < 0 means return all substrings.
func (re *Regexp) Split(s string, n int) []string {
	if n == 0 {
		return nil
	}

	if n < 0 {
		n = len(s) + 1 // Enough to get all splits
	}

	matches := re.FindAllStringIndex(s, n-1)
	if matches == nil {
		return []string{s}
	}

	result := make([]string, 0, len(matches)+1)
	prev := 0

	for _, match := range matches {
		result = append(result, s[prev:match[0]])
		prev = match[1]
	}

	// Append remaining text
	result = append(result, s[prev:])
	return result
}
