package gore

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Parser parses a regex string into an AST.
type Parser struct {
	input string
	pos   int
	// State for capturing groups
	captures int
	names    map[string]int
	flags    parseFlags
}

type parseFlags struct {
	caseInsensitive bool
}

func NewParser(input string) *Parser {
	return &Parser{
		input: input,
		names: make(map[string]int),
	}
}

func (p *Parser) Parse() (Node, error) {
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.input) {
		return nil, fmt.Errorf("unexpected character at %d: %q", p.pos, p.peek())
	}
	return node, nil
}

// parseExpr handles alternation: term | term
func (p *Parser) parseExpr() (Node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	if p.pos < len(p.input) && p.peek() == '|' {
		p.consume() // eat |
		right, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		// Merge specific logic if recursive right is already Alternate?
		// For simplicity, just binary tree or append if safe.
		// Standard optimization is to flatten Alternates.
		if alt, ok := right.(*Alternate); ok {
			return &Alternate{Nodes: append([]Node{left}, alt.Nodes...)}, nil
		}
		return &Alternate{Nodes: []Node{left, right}}, nil
	}
	return left, nil
}

// parseTerm handles concatenation: factor factor
func (p *Parser) parseTerm() (Node, error) {
	var nodes []Node
	for p.pos < len(p.input) && p.peek() != '|' && p.peek() != ')' {
		node, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return &Concat{Nodes: nodes}, nil
}

// parseFactor handles quantifiers: atom*, atom+, atom?
func (p *Parser) parseFactor() (Node, error) {
	atom, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	if p.pos >= len(p.input) {
		return atom, nil
	}

	ch := p.peek()
	switch ch {
	case '*', '+', '?':
		p.consume()
		q := &Quantifier{Body: atom, Greedy: true}
		if ch == '*' {
			q.Min, q.Max = 0, -1
		} else if ch == '+' {
			q.Min, q.Max = 1, -1
		} else {
			q.Min, q.Max = 0, 1
		}
		if p.pos < len(p.input) && p.peek() == '?' {
			p.consume()
			q.Greedy = false
		}
		return q, nil
	case '{':
		p.consume() // eat {

		// Parse minimum
		minStr := ""
		for p.pos < len(p.input) && p.peek() >= '0' && p.peek() <= '9' {
			minStr += string(p.consume())
		}
		if minStr == "" {
			return nil, fmt.Errorf("invalid quantifier: missing number")
		}
		min, err := strconv.Atoi(minStr)
		if err != nil {
			return nil, fmt.Errorf("invalid quantifier: %v", err)
		}

		max := min // Default: exactly n

		if p.pos < len(p.input) && p.peek() == ',' {
			p.consume() // eat ,

			if p.pos < len(p.input) && p.peek() == '}' {
				// {n,} means n or more
				max = -1
			} else {
				// {n,m} means n to m
				maxStr := ""
				for p.pos < len(p.input) && p.peek() >= '0' && p.peek() <= '9' {
					maxStr += string(p.consume())
				}
				if maxStr == "" {
					return nil, fmt.Errorf("invalid quantifier: missing max")
				}
				max, err = strconv.Atoi(maxStr)
				if err != nil {
					return nil, fmt.Errorf("invalid quantifier: %v", err)
				}
			}
		}

		if p.pos >= len(p.input) || p.consume() != '}' {
			return nil, fmt.Errorf("unclosed quantifier")
		}

		q := &Quantifier{Body: atom, Min: min, Max: max, Greedy: true}

		// Check for non-greedy modifier
		if p.pos < len(p.input) && p.peek() == '?' {
			p.consume()
			q.Greedy = false
		}

		return q, nil
	}
	return atom, nil
}

// parseAtom handles literals, groups, char classes
func (p *Parser) parseAtom() (Node, error) {
	ch := p.peek()
	switch ch {
	case '(':
		p.consume()
		return p.parseGroup()
	case '[':
		p.consume()
		return p.parseCharClass()
	case '.':
		p.consume()
		// . matches anything but newline
		// Allow Dot to be a CharClass for everything except \n
		// For now using negated CharClass [\n]
		return &CharClass{Negated: true, Ranges: []RuneRange{{Lo: '\n', Hi: '\n'}}}, nil

	case '\\':
		p.consume() // eat \
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("trailing backslash")
		}
		esc := p.consume()
		switch esc {
		// Character classes
		case 'd':
			return &CharClass{Ranges: []RuneRange{{'0', '9'}}, FoldCase: p.flags.caseInsensitive}, nil
		case 'D':
			return &CharClass{Ranges: []RuneRange{{'0', '9'}}, Negated: true, FoldCase: p.flags.caseInsensitive}, nil
		case 'w':
			return &CharClass{Ranges: []RuneRange{{'0', '9'}, {'A', 'Z'}, {'_', '_'}, {'a', 'z'}}, FoldCase: p.flags.caseInsensitive}, nil
		case 'W':
			return &CharClass{Ranges: []RuneRange{{'0', '9'}, {'A', 'Z'}, {'_', '_'}, {'a', 'z'}}, Negated: true, FoldCase: p.flags.caseInsensitive}, nil
		case 's':
			return &CharClass{Ranges: []RuneRange{{'\t', '\t'}, {'\n', '\n'}, {'\r', '\r'}, {' ', ' '}}, FoldCase: p.flags.caseInsensitive}, nil
		case 'S':
			return &CharClass{Ranges: []RuneRange{{'\t', '\t'}, {'\n', '\n'}, {'\r', '\r'}, {' ', ' '}}, Negated: true, FoldCase: p.flags.caseInsensitive}, nil

		// Assertions (no fold)
		case 'b':
			return &Assertion{Kind: AssertWordBoundary}, nil
		case 'B':
			return &Assertion{Kind: AssertNotWordBoundary}, nil

		// Literal escapes
		case 'n':
			return &Literal{Runes: []rune{'\n'}, FoldCase: p.flags.caseInsensitive}, nil
		case 't':
			return &Literal{Runes: []rune{'\t'}, FoldCase: p.flags.caseInsensitive}, nil
		case 'r':
			return &Literal{Runes: []rune{'\r'}, FoldCase: p.flags.caseInsensitive}, nil
		case 'f':
			return &Literal{Runes: []rune{'\f'}, FoldCase: p.flags.caseInsensitive}, nil
		case 'v':
			return &Literal{Runes: []rune{'\v'}, FoldCase: p.flags.caseInsensitive}, nil

		// Escaped metacharacters
		case '.', '*', '+', '?', '|', '(', ')', '[', ']', '{', '}', '^', '$', '\\':
			return &Literal{Runes: []rune{esc}, FoldCase: p.flags.caseInsensitive}, nil

		default:
			// Treat as literal
			return &Literal{Runes: []rune{esc}, FoldCase: p.flags.caseInsensitive}, nil
		}
	case '^':
		p.consume()
		return &Assertion{Kind: AssertStartText}, nil
	case '$':
		p.consume()
		return &Assertion{Kind: AssertEndText}, nil
	case '|', ')':
		return nil, fmt.Errorf("unexpected meta char: %c", ch)
	default:
		p.consume()
		return &Literal{Runes: []rune{ch}, FoldCase: p.flags.caseInsensitive}, nil
	}
}

func (p *Parser) parseCharClass() (Node, error) {
	// Already consumed [
	negated := false
	if p.peek() == '^' {
		p.consume()
		negated = true
	}

	var ranges []RuneRange

	// If ] is the first char (after optional ^), it's a literal ]
	// But standard logic is: if ] is first, it's literal.
	if p.peek() == ']' {
		p.consume()
		ranges = append(ranges, RuneRange{Lo: ']', Hi: ']'})
	}

	for p.pos < len(p.input) && p.peek() != ']' {
		r1 := p.consume_cc_char()

		// Check for range a-z
		if p.peek() == '-' {
			p.consume() // eat -
			if p.peek() == ']' {
				// literal - at end
				ranges = append(ranges, RuneRange{Lo: r1, Hi: r1})
				ranges = append(ranges, RuneRange{Lo: '-', Hi: '-'})
				break
			}
			r2 := p.consume_cc_char()
			ranges = append(ranges, RuneRange{Lo: r1, Hi: r2})
		} else {
			ranges = append(ranges, RuneRange{Lo: r1, Hi: r1})
		}
	}

	if p.pos >= len(p.input) || p.consume() != ']' {
		return nil, fmt.Errorf("unclosed character class")
	}

	return &CharClass{Ranges: ranges, Negated: negated, FoldCase: p.flags.caseInsensitive}, nil
}

func (p *Parser) consume_cc_char() rune {
	if p.peek() == '\\' {
		p.consume()
		if p.pos >= len(p.input) {
			return '\\' // Should error but gracefully return
		}
		return p.consume()
	}
	return p.consume()
}

func (p *Parser) parseGroup() (Node, error) {
	// Already consumed (
	// Check for (? extensions
	if p.peek() == '?' {
		p.consume() // eat ?

		// Check for flags: (?i) or (?-i)
		if p.pos < len(p.input) && (p.peek() == 'i' || p.peek() == '-') {
			originalFlags := p.flags // Save flags before modification

			// Handle flags
			turnOn := true
			if p.peek() == '-' {
				turnOn = false
				p.consume() // eat -
			}

			if p.pos < len(p.input) && p.peek() == 'i' {
				p.consume() // eat i
				p.flags.caseInsensitive = turnOn
			}

			if p.pos < len(p.input) && p.peek() == ')' {
				p.consume() // eat )
				// This was just a flag setting group, return Empty literal
				return &Literal{Runes: []rune{}, FoldCase: p.flags.caseInsensitive}, nil
			}

			// If we have (?-i:...) it's a non-capturing group with flags
			if p.pos < len(p.input) && p.peek() == ':' {
				p.consume()                                // eat :
				defer func() { p.flags = originalFlags }() // Restore flags after group

				body, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				if p.pos >= len(p.input) || p.consume() != ')' {
					return nil, fmt.Errorf("unclosed group")
				}
				return body, nil
			}

			return nil, fmt.Errorf("invalid flag syntax")
		}

		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("invalid group syntax")
		}

		// Map: (?P<name>...), (?:...), (?=...), (?!...), (?<=...), (?<!...)

		switch p.peek() {
		case ':': // (?: non-capturing
			p.consume()
			node, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			if p.consume() != ')' {
				return nil, fmt.Errorf("unclosed non-capturing group")
			}
			return node, nil

		case 'P': // (?P<name> named group
			p.consume()
			if p.consume() != '<' {
				return nil, fmt.Errorf("expected < in named group")
			}
			nameEnd := strings.IndexRune(p.input[p.pos:], '>')
			if nameEnd == -1 {
				return nil, fmt.Errorf("unclosed group name")
			}
			name := p.input[p.pos : p.pos+nameEnd]
			p.pos += nameEnd + 1 // skip name and >

			p.captures++
			idx := p.captures
			p.names[name] = idx

			node, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			if p.consume() != ')' {
				return nil, fmt.Errorf("unclosed named group")
			}
			return &Capture{Body: node, Index: idx, Name: name}, nil

		case '=': // (?= lookahead)
			p.consume()
			return p.parseLookaround(false, false)

		case '!': // (?! neg lookahead)
			p.consume()
			return p.parseLookaround(true, false)

		case '<': // (?<= lookbehind) or (?<! neg lookbehind)
			p.consume()
			neg := false
			if p.peek() == '!' {
				neg = true
				p.consume()
			} else if p.peek() == '=' {
				p.consume()
			} else {
				return nil, fmt.Errorf("invalid lookbehind syntax")
			}
			return p.parseLookaround(neg, true)
		default:
			return nil, fmt.Errorf("invalid group extension: ?%c", p.peek())
		}
	}

	// Normal capturing group
	p.captures++
	idx := p.captures
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.consume() != ')' {
		return nil, fmt.Errorf("unclosed capturing group")
	}
	return &Capture{Body: node, Index: idx}, nil
}

func (p *Parser) parseLookaround(negative, behind bool) (Node, error) {
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.consume() != ')' {
		return nil, fmt.Errorf("unclosed lookaround")
	}
	return &Lookaround{Body: node, Negative: negative, Behind: behind}, nil
}

// Helpers

func (p *Parser) peek() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	return r
}

func (p *Parser) consume() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	r, w := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += w
	return r
}
