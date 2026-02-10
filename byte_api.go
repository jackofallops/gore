package gore

// Find returns a slice holding the text of the leftmost match in b of the regular expression.
// A return value of nil indicates no match.
func (re *Regexp) Find(b []byte) []byte {
	match := re.FindIndex(b)
	if match == nil {
		return nil
	}
	return b[match[0]:match[1]]
}

// FindIndex returns a two-element slice of integers defining the location of
// the leftmost match in b of the regular expression. A return value of nil
// indicates no match.
func (re *Regexp) FindIndex(b []byte) []int {
	return re.FindStringIndex(string(b))
}

// FindSubmatch returns a slice of slices holding the text of the leftmost match
// of the regular expression in b and the matches, if any, of its subexpressions.
// A return value of nil indicates no match.
func (re *Regexp) FindSubmatch(b []byte) [][]byte {
	submatches := re.FindStringSubmatch(string(b))
	if submatches == nil {
		return nil
	}

	result := make([][]byte, len(submatches))
	for i, s := range submatches {
		if s != "" {
			result[i] = []byte(s)
		}
	}
	return result
}

// FindAll returns a slice of all successive matches of the expression.
// A return value of nil indicates no match.
// n < 0 means return all matches.
func (re *Regexp) FindAll(b []byte, n int) [][]byte {
	indices := re.FindAllIndex(b, n)
	if indices == nil {
		return nil
	}

	result := make([][]byte, len(indices))
	for i, match := range indices {
		result[i] = b[match[0]:match[1]]
	}
	return result
}

// FindAllIndex returns a slice of all successive matches of the expression,
// as two-element slices of integers. n < 0 means return all matches.
func (re *Regexp) FindAllIndex(b []byte, n int) [][]int {
	return re.FindAllStringIndex(string(b), n)
}

// FindAllSubmatch returns a slice of all successive matches of the expression,
// as defined by FindSubmatch. n < 0 means return all matches.
func (re *Regexp) FindAllSubmatch(b []byte, n int) [][][]byte {
	allMatches := re.FindAllStringSubmatch(string(b), n)
	if allMatches == nil {
		return nil
	}

	result := make([][][]byte, len(allMatches))
	for i, match := range allMatches {
		result[i] = make([][]byte, len(match))
		for j, s := range match {
			if s != "" {
				result[i][j] = []byte(s)
			}
		}
	}
	return result
}

// Match reports whether the byte slice b contains any match of the regular expression re.
func (re *Regexp) Match(b []byte) bool {
	return re.MatchString(string(b))
}
