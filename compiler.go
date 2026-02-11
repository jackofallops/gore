package gore

// Compiler compiles an AST into a VM Program.
type Compiler struct {
	insts []Inst
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(node Node, numCaptures int) (*Prog, error) {
	c.insts = nil // reset

	// Implicit Capture Group 0 (Whole Match)
	// Save(0) -> Body -> Save(1) -> Match

	c.emit(Inst{Op: OpSave, Idx: 0})
	c.compileNode(node)
	c.emit(Inst{Op: OpSave, Idx: 1})
	start := 0 // Start is always 0 now

	c.emit(Inst{Op: OpMatch})

	prog := &Prog{
		Insts:             c.insts,
		Start:             start,
		NumCap:            numCaptures + 1, // +1 for implicit group 0
		LookbehindLengths: make(map[int]int),
	}

	// Analyze pattern for optimizations
	prog.Prefix = c.analyzePrefix(node)
	c.analyzeLookbehinds(prog)

	return prog, nil
}

// analyzePrefix extracts a literal prefix from the pattern for fast searching
func (c *Compiler) analyzePrefix(node Node) string {
	switch n := node.(type) {
	case *Literal:
		// Return literal as prefix (only if case-sensitive)
		if n.FoldCase {
			return ""
		}
		return string(n.Runes)
	case *Concat:
		// First node of concat could be prefix
		if len(n.Nodes) > 0 {
			return c.analyzePrefix(n.Nodes[0])
		}
	case *Capture:
		// Look inside capture
		return c.analyzePrefix(n.Body)
	}
	return ""
}

// analyzeLookbehinds finds fixed-length lookbehind patterns
func (c *Compiler) analyzeLookbehinds(prog *Prog) {
	for pc, inst := range prog.Insts {
		if inst.Op == OpLookaround && inst.LookBehind {
			// Analyze the lookbehind subprogram
			length := c.analyzeFixedLength(inst.Prog, inst.Prog.Start, 0)
			prog.LookbehindLengths[pc] = length
		}
	}
}

// analyzeFixedLength determines if a pattern has fixed length
// Returns length if fixed, 0 if variable
func (c *Compiler) analyzeFixedLength(prog *Prog, pc int, currentLen int) int {
	visited := make(map[int]bool)
	return c.analyzeFixedLengthRec(prog, pc, currentLen, visited)
}

func (c *Compiler) analyzeFixedLengthRec(prog *Prog, pc int, currentLen int, visited map[int]bool) int {
	if visited[pc] || pc >= len(prog.Insts) {
		return 0 // Cycle or invalid = variable length
	}
	visited[pc] = true

	inst := prog.Insts[pc]
	switch inst.Op {
	case OpMatch:
		return currentLen

	case OpChar:
		// Single rune = 1 byte typically, but could be multi-byte
		// For simplicity, count as 1 rune
		return c.analyzeFixedLengthRec(prog, pc+1, currentLen+1, visited)

	case OpCharClass, OpAny:
		return c.analyzeFixedLengthRec(prog, pc+1, currentLen+1, visited)

	case OpJmp:
		return c.analyzeFixedLengthRec(prog, inst.Out, currentLen, visited)

	case OpSplit:
		// Both branches must have same length
		len1 := c.analyzeFixedLengthRec(prog, inst.Out, currentLen, visited)
		len2 := c.analyzeFixedLengthRec(prog, inst.Out1, currentLen, visited)
		if len1 == len2 && len1 > 0 {
			return len1
		}
		return 0 // Variable length

	case OpSave:
		return c.analyzeFixedLengthRec(prog, pc+1, currentLen, visited)

	case OpAssert:
		return c.analyzeFixedLengthRec(prog, pc+1, currentLen, visited)

	default:
		return 0 // Unknown = variable
	}
}

func (c *Compiler) emit(i Inst) int {
	c.insts = append(c.insts, i)
	return len(c.insts) - 1
}

func (c *Compiler) compileNode(node Node) int {
	switch n := node.(type) {
	case *Literal:
		start := -1
		for i, r := range n.Runes {
			idx := c.emit(Inst{
				Op:       OpChar,
				Val:      r,
				FoldCase: n.FoldCase,
			})
			if i == 0 {
				start = idx
			}
		}
		return start

	case *CharClass:
		return c.emit(Inst{
			Op:       OpCharClass,
			Ranges:   n.Ranges,
			Negated:  n.Negated,
			FoldCase: n.FoldCase,
		})

	case *Concat:
		if len(n.Nodes) == 0 {
			return -1
		}
		start := c.compileNode(n.Nodes[0])
		for i := 1; i < len(n.Nodes); i++ {
			c.compileNode(n.Nodes[i])
		}
		return start

	case *Alternate:
		if len(n.Nodes) == 0 {
			return -1
		}
		if len(n.Nodes) == 1 {
			return c.compileNode(n.Nodes[0])
		}

		left := n.Nodes[0]

		var right Node
		if len(n.Nodes) == 2 {
			right = n.Nodes[1]
		} else {
			right = &Alternate{Nodes: n.Nodes[1:]}
		}

		splitIdx := c.emit(Inst{Op: OpSplit})

		c.insts[splitIdx].Out = len(c.insts)
		c.compileNode(left)

		jmpIdx := c.emit(Inst{Op: OpJmp})

		c.insts[splitIdx].Out1 = len(c.insts)
		c.compileNode(right)

		end := len(c.insts)
		c.insts[jmpIdx].Out = end

		return splitIdx

	case *Quantifier:
		return c.compileQuantifier(n)

	case *Capture:
		idx1 := c.emit(Inst{Op: OpSave, Idx: 2 * n.Index})
		c.compileNode(n.Body)
		c.emit(Inst{Op: OpSave, Idx: 2*n.Index + 1})
		return idx1

	case *Assertion:
		return c.emit(Inst{Op: OpAssert, Assert: n.Kind})

	case *Lookaround:
		subC := NewCompiler()
		subProg, _ := subC.Compile(n.Body, 0) // Lookaround captures are independent

		return c.emit(Inst{
			Op:         OpLookaround,
			Prog:       subProg,
			LookNeg:    n.Negative,
			LookBehind: n.Behind,
		})

	case *Backreference:
		return c.emit(Inst{
			Op:  OpBackref,
			Idx: n.Index,
		})
	}
	return -1
}

func (c *Compiler) compileQuantifier(q *Quantifier) int {
	start := len(c.insts)

	if q.Min == 0 && q.Max == -1 { // *
		split := c.emit(Inst{Op: OpSplit})
		c.compileNode(q.Body)
		c.emit(Inst{Op: OpJmp, Out: split})

		end := len(c.insts)
		if q.Greedy {
			c.insts[split].Out = start + 1
			c.insts[split].Out1 = end
		} else {
			c.insts[split].Out = end
			c.insts[split].Out1 = start + 1
		}
		return split
	}

	if q.Min == 1 && q.Max == -1 { // +
		bodyStart := c.compileNode(q.Body)
		split := c.emit(Inst{Op: OpSplit})

		end := len(c.insts)
		if q.Greedy {
			c.insts[split].Out = bodyStart
			c.insts[split].Out1 = end
		} else {
			c.insts[split].Out = end
			c.insts[split].Out1 = bodyStart
		}
		return bodyStart
	}

	if q.Min == 0 && q.Max == 1 { // ?
		split := c.emit(Inst{Op: OpSplit})
		c.compileNode(q.Body)
		end := len(c.insts)

		if q.Greedy {
			c.insts[split].Out = start + 1
			c.insts[split].Out1 = end
		} else {
			c.insts[split].Out = end
			c.insts[split].Out1 = start + 1
		}
		return split
	}

	// {n} - exactly n times
	if q.Min == q.Max && q.Max > 0 {
		for i := 0; i < q.Min; i++ {
			c.compileNode(q.Body)
		}
		return start
	}

	// {n,m} - between n and m times (inclusive)
	if q.Min >= 0 && q.Max > q.Min {
		// Required repetitions
		for i := 0; i < q.Min; i++ {
			c.compileNode(q.Body)
		}

		// Optional repetitions (max - min)
		for i := 0; i < q.Max-q.Min; i++ {
			split := c.emit(Inst{Op: OpSplit})
			bodyStart := len(c.insts)
			c.compileNode(q.Body)
			end := len(c.insts)

			if q.Greedy {
				c.insts[split].Out = bodyStart
				c.insts[split].Out1 = end
			} else {
				c.insts[split].Out = end
				c.insts[split].Out1 = bodyStart
			}
		}
		return start
	}

	// {n,} - n or more times
	if q.Min > 0 && q.Max == -1 {
		// Required repetitions
		for i := 0; i < q.Min; i++ {
			c.compileNode(q.Body)
		}

		// Then * (zero or more)
		split := c.emit(Inst{Op: OpSplit})
		bodyStart := len(c.insts)
		c.compileNode(q.Body)
		c.emit(Inst{Op: OpJmp, Out: split})
		end := len(c.insts)

		if q.Greedy {
			c.insts[split].Out = bodyStart
			c.insts[split].Out1 = end
		} else {
			c.insts[split].Out = end
			c.insts[split].Out1 = bodyStart
		}
		return start
	}

	return -1
}
