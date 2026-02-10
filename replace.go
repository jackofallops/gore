package gore

import (
	"strings"
)

// ReplaceAllString replaces all matches of the regular expression with the replacement string.
// Inside repl, $ signs are interpreted as in Expand, so for instance $1 represents
// the text of the first submatch.
func (re *Regexp) ReplaceAllString(src, repl string) string {
	allMatches := re.FindAllStringSubmatch(src, -1)
	if allMatches == nil {
		return src
	}

	indices := re.FindAllStringIndex(src, -1)
	if indices == nil || len(indices) != len(allMatches) {
		return src
	}

	var result strings.Builder
	lastEnd := 0

	for i, match := range allMatches {
		// Append text before match
		result.WriteString(src[lastEnd:indices[i][0]])

		// Expand template with captures from this match
		expanded := re.expandStringWithCaptures(repl, match)
		result.WriteString(expanded)

		lastEnd = indices[i][1]
	}

	// Append remaining text
	result.WriteString(src[lastEnd:])
	return result.String()
}

// ReplaceAllLiteralString replaces all matches with the replacement string literally
// (no template expansion).
func (re *Regexp) ReplaceAllLiteralString(src, repl string) string {
	return re.ReplaceAllStringFunc(src, func(string) string {
		return repl
	})
}

// ReplaceAllStringFunc replaces all matches using a function to generate replacement text.
func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	// Use FindAllStringSubmatchIndex to get capture positions
	input := NewStringInput(src)
	inputLen := input.Len()
	pos := 0

	var result strings.Builder
	lastEnd := 0

	for pos <= inputLen {
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
			matchStart := caps[0]
			matchEnd := caps[1]

			// Append text before match
			result.WriteString(src[lastEnd:matchStart])

			// Apply replacement function
			matchText := src[matchStart:matchEnd]
			result.WriteString(repl(matchText))

			lastEnd = matchEnd

			// Advance past match (handle zero-width)
			if matchEnd == pos {
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

	// Append remaining text
	result.WriteString(src[lastEnd:])
	return result.String()
}

// expandString expands template strings with $1, $2, $name substitutions.
// This is used by ReplaceAllString.
func (re *Regexp) expandString(template, src, match string) string {
	// We need to re-match to get the submatches
	// Find where this match occurs in src to get proper captures
	input := NewStringInput(src)
	vm := NewVM(re.prog, input)

	// Find the match position
	matchPos := strings.Index(src, match)
	if matchPos == -1 {
		return template
	}

	// Run VM at that position to get captures
	matched, caps := vm.Run(matchPos)
	if !matched || len(caps) < 2 {
		return template
	}

	// Build result array from captures
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
			result[i] = src[start:end]
		}
	}

	var expanded strings.Builder
	i := 0
	for i < len(template) {
		if template[i] != '$' {
			expanded.WriteByte(template[i])
			i++
			continue
		}

		// Found $
		i++
		if i >= len(template) {
			expanded.WriteByte('$')
			break
		}

		// Handle $$
		if template[i] == '$' {
			expanded.WriteByte('$')
			i++
			continue
		}

		// Handle ${name} or ${1}
		if template[i] == '{' {
			i++
			nameStart := i
			for i < len(template) && template[i] != '}' {
				i++
			}
			if i >= len(template) {
				// Unclosed ${, treat as literal
				expanded.WriteString("${")
				i = nameStart
				continue
			}
			name := template[nameStart:i]
			i++ // skip }

			// Try numeric first
			if name >= "0" && name <= "9" {
				idx := int(name[0] - '0')
				if idx < len(result) && result[idx] != "" {
					expanded.WriteString(result[idx])
				}
			} else {
				// Named group
				idx := re.SubexpIndex(name)
				if idx >= 0 && idx < len(result) && result[idx] != "" {
					expanded.WriteString(result[idx])
				}
			}
			continue
		}

		// Handle $1, $2, ... $9
		if template[i] >= '0' && template[i] <= '9' {
			idx := int(template[i] - '0')
			if idx < len(result) && result[idx] != "" {
				expanded.WriteString(result[idx])
			}
			i++
			continue
		}

		// Handle $name (alphanumeric identifier)
		nameStart := i
		for i < len(template) && isIdentChar(template[i]) {
			i++
		}
		if i > nameStart {
			name := template[nameStart:i]
			idx := re.SubexpIndex(name)
			if idx >= 0 && idx < len(result) && result[idx] != "" {
				expanded.WriteString(result[idx])
			}
			continue
		}

		// Invalid $, treat as literal
		expanded.WriteByte('$')
	}

	return expanded.String()
}

// isIdentChar returns true if c is a valid identifier character (letter, digit, underscore).
func isIdentChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

// expandStringWithCaptures expands template with pre-extracted captures array.
func (re *Regexp) expandStringWithCaptures(template string, captures []string) string {
	var expanded strings.Builder
	i := 0
	for i < len(template) {
		if template[i] != '$' {
			expanded.WriteByte(template[i])
			i++
			continue
		}

		// Found $
		i++
		if i >= len(template) {
			expanded.WriteByte('$')
			break
		}

		// Handle $$
		if template[i] == '$' {
			expanded.WriteByte('$')
			i++
			continue
		}

		// Handle ${name} or ${1}
		if template[i] == '{' {
			i++
			nameStart := i
			for i < len(template) && template[i] != '}' {
				i++
			}
			if i >= len(template) {
				// Unclosed ${, treat as literal
				expanded.WriteString("${")
				i = nameStart
				continue
			}
			name := template[nameStart:i]
			i++ // skip }

			// Try numeric first
			if name >= "0" && name <= "9" {
				idx := int(name[0] - '0')
				if idx < len(captures) && captures[idx] != "" {
					expanded.WriteString(captures[idx])
				}
			} else {
				// Named group
				idx := re.SubexpIndex(name)
				if idx >= 0 && idx < len(captures) && captures[idx] != "" {
					expanded.WriteString(captures[idx])
				}
			}
			continue
		}

		// Handle $1, $2, ... $9
		if template[i] >= '0' && template[i] <= '9' {
			idx := int(template[i] - '0')
			if idx < len(captures) && captures[idx] != "" {
				expanded.WriteString(captures[idx])
			}
			i++
			continue
		}

		// Handle $name (alphanumeric identifier)
		nameStart := i
		for i < len(template) && isIdentChar(template[i]) {
			i++
		}
		if i > nameStart {
			name := template[nameStart:i]
			idx := re.SubexpIndex(name)
			if idx >= 0 && idx < len(captures) && captures[idx] != "" {
				expanded.WriteString(captures[idx])
			}
			continue
		}

		// Invalid $, treat as literal
		expanded.WriteByte('$')
	}

	return expanded.String()
}

// ReplaceAll replaces all matches in a byte slice.
func (re *Regexp) ReplaceAll(src, repl []byte) []byte {
	return []byte(re.ReplaceAllString(string(src), string(repl)))
}

// ReplaceAllLiteral replaces all matches in a byte slice literally.
func (re *Regexp) ReplaceAllLiteral(src, repl []byte) []byte {
	return []byte(re.ReplaceAllLiteralString(string(src), string(repl)))
}

// ReplaceAllFunc replaces all matches in a byte slice using a function.
func (re *Regexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	return []byte(re.ReplaceAllStringFunc(string(src), func(s string) string {
		return string(repl([]byte(s)))
	}))
}
