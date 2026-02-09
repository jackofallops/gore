package gore

import (
	"fmt"
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
		// TODO: Implement {n,m} parsing
		return atom, nil
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
		case 'd':
			return &CharClass{Ranges: []RuneRange{{'0', '9'}}}, nil
		case 'w':
			return &CharClass{Ranges: []RuneRange{{'0', '9'}, {'A', 'Z'}, {'_', '_'}, {'a', 'z'}}}, nil
		case 's':
			return &CharClass{Ranges: []RuneRange{{'\t', '\t'}, {'\n', '\n'}, {'\r', '\r'}, {' ', ' '}}}, nil
		default:
			return &Literal{Runes: []rune{esc}}, nil
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
		return &Literal{Runes: []rune{ch}}, nil
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

	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unclosed character class")
	}
	p.consume() // eat ]

	return &CharClass{Ranges: ranges, Negated: negated}, nil
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
