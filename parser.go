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
	multiline       bool
	dotall          bool // for future (?s) implementation
}

func NewParser(input string) *Parser {
	return &Parser{
		input: input,
		names: make(map[string]int),
	}
}

// isIdentStart returns true if r is a valid identifier start character (letter or underscore).
func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

// isIdentRune returns true if r is a valid identifier character (letter, digit, underscore).
func isIdentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
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
		switch ch {
		case '*':
			q.Min, q.Max = 0, -1
		case '+':
			q.Min, q.Max = 1, -1
		default: // '?'
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
				// Validate min <= max
				if min > max {
					return nil, fmt.Errorf("invalid quantifier {%d,%d}: min cannot be greater than max", min, max)
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
		if p.flags.dotall {
			// Dotall mode: . matches any character including \n
			// Match all Unicode characters
			return &CharClass{
				Negated: false,
				Ranges:  []RuneRange{{Lo: 0, Hi: '\U0010FFFF'}},
			}, nil
		}
		// Default: . matches anything but newline
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
		case 'A':
			return &Assertion{Kind: AssertStringStart}, nil
		case 'Z':
			return &Assertion{Kind: AssertStringEnd}, nil
		case 'z':
			return &Assertion{Kind: AssertAbsoluteEnd}, nil

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
			// Check for backreference \1, \2, etc.
			if esc >= '1' && esc <= '9' {
				return &Backreference{Index: int(esc - '0')}, nil
			}
			// Treat as literal
			return &Literal{Runes: []rune{esc}, FoldCase: p.flags.caseInsensitive}, nil
		}
	case '^':
		p.consume()
		return &Assertion{Kind: AssertStartText, Multiline: p.flags.multiline}, nil
	case '$':
		p.consume()
		return &Assertion{Kind: AssertEndText, Multiline: p.flags.multiline}, nil
	case '|', ')':
		return nil, fmt.Errorf("unexpected meta char: %c", ch)
	default:
		// Check for quantifier metacharacters without target
		if ch == '*' || ch == '+' || ch == '?' || ch == '{' {
			return nil, fmt.Errorf("quantifier %q requires a target", ch)
		}
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
		// Check for escape sequences that expand to multiple ranges
		if p.peek() == '\\' && p.pos+1 < len(p.input) {
			nextChar := p.input[p.pos+1]
			switch nextChar {
			case 'd':
				p.consume() // eat \
				p.consume() // eat d
				ranges = append(ranges, RuneRange{Lo: '0', Hi: '9'})
				continue
			case 'D':
				// \D inside [] means NOT digit, but we can't easily handle negation inside class
				// For now, treat as error or expand to many ranges
				return nil, fmt.Errorf("\\D not supported inside character class")
			case 'w':
				p.consume() // eat \
				p.consume() // eat w
				ranges = append(ranges, RuneRange{Lo: '0', Hi: '9'})
				ranges = append(ranges, RuneRange{Lo: 'A', Hi: 'Z'})
				ranges = append(ranges, RuneRange{Lo: '_', Hi: '_'})
				ranges = append(ranges, RuneRange{Lo: 'a', Hi: 'z'})
				continue
			case 'W':
				// \W inside [] is problematic
				return nil, fmt.Errorf("\\W not supported inside character class")
			case 's':
				p.consume() // eat \
				p.consume() // eat s
				ranges = append(ranges, RuneRange{Lo: '\t', Hi: '\t'})
				ranges = append(ranges, RuneRange{Lo: '\n', Hi: '\n'})
				ranges = append(ranges, RuneRange{Lo: '\r', Hi: '\r'})
				ranges = append(ranges, RuneRange{Lo: ' ', Hi: ' '})
				continue
			case 'S':
				// \S inside [] is problematic
				return nil, fmt.Errorf("\\S not supported inside character class")
			}
		}

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
			// Validate that Lo <= Hi
			if r1 > r2 {
				return nil, fmt.Errorf("invalid character class range: %c-%c (start > end)", r1, r2)
			}
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
		esc := p.consume()
		// Handle common escape sequences
		switch esc {
		case 'n':
			return '\n'
		case 't':
			return '\t'
		case 'r':
			return '\r'
		case 'f':
			return '\f'
		case 'v':
			return '\v'
		default:
			// For other escapes, return the literal character
			return esc
		}
	}
	return p.consume()
}

func (p *Parser) parseGroup() (Node, error) {
	// Already consumed (
	// Check for (? extensions
	if p.peek() == '?' {
		p.consume() // eat ?

		// Check for flags: (?i) (?m) (?s) or combinations (?im) (?-i)
		if p.pos < len(p.input) && (p.peek() == 'i' || p.peek() == 'm' ||
			p.peek() == 's' || p.peek() == '-') {
			originalFlags := p.flags // Save flags before modification

			turnOn := true
			for p.pos < len(p.input) {
				ch := p.peek()
				if ch == ')' || ch == ':' {
					break
				}

				if ch == '-' {
					turnOn = false
					p.consume()
					continue
				}

				switch ch {
				case 'i':
					p.consume()
					p.flags.caseInsensitive = turnOn
				case 'm':
					p.consume()
					p.flags.multiline = turnOn
				case 's':
					p.consume()
					p.flags.dotall = turnOn
				default:
					return nil, fmt.Errorf("unknown flag: %c", ch)
				}
			}

			// Handle (?flags) vs (?flags:...)
			if p.pos < len(p.input) && p.peek() == ')' {
				p.consume()
				// This was just a flag setting group, return Empty literal
				return &Literal{Runes: []rune{}, FoldCase: p.flags.caseInsensitive}, nil
			}

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

			// Validate name is not empty
			if name == "" {
				return nil, fmt.Errorf("empty capture group name")
			}

			// Validate name starts with letter or underscore
			firstChar := rune(name[0])
			if !isIdentStart(firstChar) {
				return nil, fmt.Errorf("invalid capture group name %q: must start with letter or underscore", name)
			}

			// Validate name contains only alphanumeric and underscore
			for _, ch := range name {
				if !isIdentRune(ch) {
					return nil, fmt.Errorf("invalid capture group name %q: contains invalid character %q", name, ch)
				}
			}

			// Check for duplicate names
			if existingIdx, exists := p.names[name]; exists {
				return nil, fmt.Errorf("duplicate capture group name %q (already used for group %d)", name, existingIdx)
			}

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
