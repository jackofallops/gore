package gore

// VM executes the regex program.
type VM struct {
	prog  *Prog
	input Input
}

func NewVM(prog *Prog, input Input) *VM {
	return &VM{prog: prog, input: input}
}

// Run executes the VM starting at the given position.
// Returns true if match found, and the capture positions.
func (vm *VM) Run(pos int) (bool, []int) {
	// Cap slice size: 2 * numCap
	caps := make([]int, vm.prog.NumCap*2)
	for i := range caps {
		caps[i] = -1
	}

	if vm.match(vm.prog.Start, pos, caps) {
		return true, caps
	}
	return false, nil
}

// match is the core backtracking function.
func (vm *VM) match(pc int, pos int, caps []int) bool {
	// Infinite loop protection could be added here (step limit)

	for {
		if pc >= len(vm.prog.Insts) {
			return false
		}
		inst := vm.prog.Insts[pc]

		switch inst.Op {
		case OpMatch:
			return true

		case OpChar:
			r, w := vm.input.Step(pos)
			if r != inst.Val {
				return false
			}
			pos += w
			pc++

		case OpCharClass:
			r, w := vm.input.Step(pos)
			// CharClass must consume a character.
			if w == 0 { // EOF
				return false
			}
			if !matchClass(r, inst.Ranges, inst.Negated) {
				return false
			}
			pos += w
			pc++

		case OpAny:
			r, w := vm.input.Step(pos)
			if r == 0 && w == 0 { // EOF
				return false
			}
			// Dot usually doesn't match newline.
			// If we want DotAll, we need a flag. Assuming not for now.
			if r == '\n' {
				return false
			}
			pos += w
			pc++

		case OpJmp:
			pc = inst.Out

		case OpSplit:
			// Backtracking split
			// Try Out first (greedy default ordering in compiler determines this)
			// Copy matches (captures)
			// Efficient capture copy? For now full slice copy.

			// Branch 1
			capsCopy := make([]int, len(caps))
			copy(capsCopy, caps)
			if vm.match(inst.Out, pos, capsCopy) {
				copy(caps, capsCopy)
				return true
			}

			// Branch 2
			return vm.match(inst.Out1, pos, caps)

		case OpSave:
			caps[inst.Idx] = pos
			pc++

		case OpAssert:
			if !vm.checkAssertion(inst.Assert, pos) {
				return false
			}
			pc++

		case OpLookaround:
			subVM := NewVM(inst.Prog, vm.input)

			matched := false

			if inst.LookBehind {
				// Naive Lookbehind:
				// Search for a match ENDING at current `pos`.
				// Try matching from i = 0 to pos.
				// This is O(pos) scans. Very slow but correct for variable length.
				// TODO: Optimize with reverse matching or length constraints.

				for i := 0; i <= pos; i++ {
					// Hack: Use RunWithEnd
					end, ok := subVM.runWithEnd(i)
					if ok && end == pos {
						matched = true
						break
					}
				}

			} else {
				// Lookahead
				matched, _ = subVM.Run(pos) // Run sub-program at current pos
			}

			if inst.LookNeg {
				if matched {
					return false
				}
			} else {
				if !matched {
					return false
				}
			}
			pc++ // Success, continue without consuming input
		}
	}
}

// runWithEnd is a helper to get end position of match
// This duplicates logic of Run/match but returns end pos.
// Ideally refactor Run/match to return int (end pos or -1)
func (vm *VM) runWithEnd(pos int) (int, bool) {
	return vm.recMatch(vm.prog.Start, pos, make([]int, vm.prog.NumCap*2))
}

func (vm *VM) recMatch(pc int, pos int, caps []int) (int, bool) {
	for {
		if pc >= len(vm.prog.Insts) {
			return -1, false
		}
		inst := vm.prog.Insts[pc]

		switch inst.Op {
		case OpMatch:
			return pos, true // Return current pos

		case OpChar:
			r, w := vm.input.Step(pos)
			if r != inst.Val {
				return -1, false
			}
			pos += w
			pc++

		case OpCharClass:
			r, w := vm.input.Step(pos)
			if w == 0 {
				return -1, false
			} // EOF check
			if !matchClass(r, inst.Ranges, inst.Negated) {
				return -1, false
			}
			pos += w
			pc++

		case OpAny:
			r, w := vm.input.Step(pos)
			if r == 0 && w == 0 {
				return -1, false
			}
			if r == '\n' {
				return -1, false
			}
			pos += w
			pc++

		case OpJmp:
			pc = inst.Out

		case OpSplit:
			caps1 := make([]int, len(caps))
			copy(caps1, caps)
			if end, ok := vm.recMatch(inst.Out, pos, caps1); ok {
				return end, true
			}
			return vm.recMatch(inst.Out1, pos, caps) // Tail call

		case OpSave:
			caps[inst.Idx] = pos
			pc++

		case OpAssert:
			if !vm.checkAssertion(inst.Assert, pos) {
				return -1, false
			}
			pc++

		case OpLookaround:
			// Recursive check
			subVM := NewVM(inst.Prog, vm.input)
			if inst.LookBehind {
				matched := false
				for i := 0; i <= pos; i++ {
					if end, ok := subVM.runWithEnd(i); ok && end == pos {
						matched = true
						break
					}
				}
				if inst.LookNeg {
					if matched {
						return -1, false
					}
				} else {
					if !matched {
						return -1, false
					}
				}
			} else {
				// Lookahead
				_, ok := subVM.runWithEnd(pos)
				if inst.LookNeg {
					if ok {
						return -1, false
					}
				} else {
					if !ok {
						return -1, false
					}
				}
			}
			pc++
		}
	}
}

func matchClass(r rune, ranges []RuneRange, negated bool) bool {
	matched := false
	for _, rng := range ranges {
		if r >= rng.Lo && r <= rng.Hi {
			matched = true
			break
		}
	}
	if negated {
		return !matched
	}
	return matched
}

func (vm *VM) checkAssertion(kind AssertionType, pos int) bool {
	switch kind {
	case AssertStartText:
		return pos == 0
	case AssertEndText:
		r, _ := vm.input.Step(pos)
		return r == 0 // EOF
	}
	return true
}
