package gore

// Compiler compiles an AST into a VM Program.
type Compiler struct {
	insts []Inst
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(node Node) (*Prog, error) {
	c.insts = nil // reset

	// Implicit Capture Group 0 (Whole Match)
	// Save(0) -> Body -> Save(1) -> Match

	c.emit(Inst{Op: OpSave, Idx: 0})
	c.compileNode(node)
	c.emit(Inst{Op: OpSave, Idx: 1})
	start := 0 // Start is always 0 now

	c.emit(Inst{Op: OpMatch})

	return &Prog{
		Insts:  c.insts,
		Start:  start,
		NumCap: 10, // TODO: Count captures dynamically or pass from parser
	}, nil
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
			idx := c.emit(Inst{Op: OpChar, Val: r})
			if i == 0 {
				start = idx
			}
		}
		return start

	case *CharClass:
		return c.emit(Inst{Op: OpCharClass, Ranges: n.Ranges, Negated: n.Negated})

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
		subProg, _ := subC.Compile(n.Body)
		// Note: subProg currently has implicit Capture 0 too.

		return c.emit(Inst{
			Op:         OpLookaround,
			Prog:       subProg,
			LookNeg:    n.Negative,
			LookBehind: n.Behind,
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

	return -1
}
