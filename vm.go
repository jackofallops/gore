package gore

import (
	"sync"
	"unicode"
)

// Pool for capture slice allocations to reduce GC pressure
var capsPool = sync.Pool{
	New: func() any {
		// Pre-allocate reasonable size
		// Return a pointer to slice to avoid allocation when putting into interface{}
		s := make([]int, 0, 20)
		return &s
	},
}

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
	// Get caps from pool and ensure proper size
	poolCapsPtr := capsPool.Get().(*[]int)
	caps := (*poolCapsPtr)[:0] // Reset length

	// Ensure capacity
	needed := vm.prog.NumCap * 2
	if cap(caps) < needed {
		caps = make([]int, needed)
	} else {
		caps = caps[:needed]
	}

	// Initialize to -1
	for i := range caps {
		caps[i] = -1
	}

	endPos, matched := vm.match(vm.prog.Start, pos, caps)
	if matched {
		// Return actual caps, don't put back in pool since caller uses them
		// endPos not used here but needed for match signature consistency
		_ = endPos
		return true, caps
	}

	// Return caps to pool
	// Return caps to pool
	// Update pointer to point to potentially new slice (if realloc)
	*poolCapsPtr = caps
	capsPool.Put(poolCapsPtr)
	return false, nil
}

// match is the unified backtracking function.
// Returns (endPos, matched) where endPos is the position after match.
func (vm *VM) match(pc int, pos int, caps []int) (int, bool) {
	// Iteration limit to prevent infinite loops
	const maxSteps = 1000000
	steps := 0

	for {
		steps++
		if steps > maxSteps || pc >= len(vm.prog.Insts) {
			return -1, false
		}

		inst := vm.prog.Insts[pc]

		switch inst.Op {
		case OpMatch:
			return pos, true

		case OpChar:
			r, w := vm.input.Step(pos)
			matched := false
			if inst.FoldCase {
				matched = simpleFoldEqual(r, inst.Val)
			} else {
				matched = r == inst.Val
			}
			if !matched {
				return -1, false
			}
			pos += w
			pc++

		case OpCharClass:
			r, w := vm.input.Step(pos)
			if w == 0 { // EOF
				return -1, false
			}
			if !matchClass(r, inst.Ranges, inst.Negated, inst.FoldCase) {
				return -1, false
			}
			pos += w
			pc++

		case OpAny:
			r, w := vm.input.Step(pos)
			if w == 0 { // EOF
				return -1, false
			}
			if r == '\n' { // Dot doesn't match newline
				return -1, false
			}
			pos += w
			pc++

		case OpJmp:
			pc = inst.Out

		case OpSplit:
			// Backtracking split: try both branches
			// Get caps copy from pool
			poolCapsPtr := capsPool.Get().(*[]int)
			capsCopy := (*poolCapsPtr)[:0]
			if cap(capsCopy) < len(caps) {
				capsCopy = make([]int, len(caps))
			} else {
				capsCopy = capsCopy[:len(caps)]
			}
			copy(capsCopy, caps)

			// Try first branch
			if endPos, ok := vm.match(inst.Out, pos, capsCopy); ok {
				copy(caps, capsCopy)
				*poolCapsPtr = capsCopy
				capsPool.Put(poolCapsPtr)
				return endPos, true
			}

			// Return copy to pool
			// Return copy to pool
			*poolCapsPtr = capsCopy
			capsPool.Put(poolCapsPtr)

			// Try second branch (tail call optimization possible)
			return vm.match(inst.Out1, pos, caps)

		case OpSave:
			caps[inst.Idx] = pos
			pc++

		case OpAssert:
			if !vm.checkAssertion(inst.Assert, pos) {
				return -1, false
			}
			pc++

		case OpLookaround:
			subVM := NewVM(inst.Prog, vm.input)
			matched := false

			if inst.LookBehind {
				// Check if this is a fixed-length lookbehind
				fixedLen, exists := vm.prog.LookbehindLengths[pc]

				if exists && fixedLen > 0 {
					// Optimized: fixed-length lookbehind O(1)
					// Only try matching from the exact position
					startPos := pos - fixedLen
					if startPos >= 0 {
						if endPos, ok := subVM.match(subVM.prog.Start, startPos, make([]int, subVM.prog.NumCap*2)); ok && endPos == pos {
							matched = true
						}
					}
				} else {
					// Fallback: O(pos) scan for variable-length lookbehind
					for i := 0; i <= pos; i++ {
						if endPos, ok := subVM.match(subVM.prog.Start, i, make([]int, subVM.prog.NumCap*2)); ok && endPos == pos {
							matched = true
							break
						}
					}
				}
			} else {
				// Lookahead
				_, matched = subVM.match(subVM.prog.Start, pos, make([]int, subVM.prog.NumCap*2))
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
			pc++

		case OpBackref:
			// Get the capture group index (1-based in AST, but we store as 1-based)
			capIdx := inst.Idx
			// Captures are stored as pairs: [start0, end0, start1, end1, ...]
			// Group 0 is the whole match, group 1 is at indices 2,3, etc.
			startIdx := capIdx * 2
			endIdx := capIdx*2 + 1

			// Check if capture group exists and has been captured
			if startIdx >= len(caps) || endIdx >= len(caps) {
				return -1, false
			}

			capStart := caps[startIdx]
			capEnd := caps[endIdx]

			// If capture group hasn't been captured yet or is empty
			if capStart == -1 || capEnd == -1 {
				// Empty backreference matches empty string
				pc++
				continue
			}

			// Match the captured text at the current position
			capLen := capEnd - capStart
			for i := 0; i < capLen; i++ {
				r1, w1 := vm.input.Step(capStart + i)
				r2, w2 := vm.input.Step(pos + i)

				// Check EOF
				if w1 == 0 || w2 == 0 {
					return -1, false
				}

				// Compare runes
				if r1 != r2 {
					return -1, false
				}
			}

			// Advance position by the length of the matched backreference
			pos += capLen
			pc++
		}
	}
}

// matchClass checks if rune r matches the character class.
// Optimized with fast-path for common single-range classes.
func matchClass(r rune, ranges []RuneRange, negated bool, foldCase bool) bool {
	matched := false

	// Case folding optimization
	if foldCase {
		// Try original rune first
		if checkRanges(r, ranges) {
			matched = true
		} else {
			// Try folded rune
			// SimpleFold iterates over unicode equivalence classes
			f := unicode.SimpleFold(r)
			for f != r {
				if checkRanges(f, ranges) {
					matched = true
					break
				}
				f = unicode.SimpleFold(f)
			}
		}
	} else {
		matched = checkRanges(r, ranges)
	}

	if negated {
		return !matched
	}
	return matched
}

// checkRanges checks if rune r is in any of the ranges
func checkRanges(r rune, ranges []RuneRange) bool {
	// Fast path for single range
	if len(ranges) == 1 {
		return r >= ranges[0].Lo && r <= ranges[0].Hi
	}

	// General case
	for _, rng := range ranges {
		if r >= rng.Lo && r <= rng.Hi {
			return true
		}
	}
	return false
}

// simpleFoldEqual checks if r1 and r2 are equal under Unicode case folding
func simpleFoldEqual(r1, r2 rune) bool {
	if r1 == r2 {
		return true
	}
	// Iterate through the fold cycle
	f := unicode.SimpleFold(r1)
	for f != r1 {
		if f == r2 {
			return true
		}
		f = unicode.SimpleFold(f)
	}
	return false
}

func (vm *VM) checkAssertion(kind AssertionType, pos int) bool {
	switch kind {
	case AssertStartText:
		return pos == 0
	case AssertEndText:
		r, _ := vm.input.Step(pos)
		return r == 0 // EOF
	case AssertWordBoundary:
		return vm.isWordBoundary(pos)
	case AssertNotWordBoundary:
		return !vm.isWordBoundary(pos)
	}
	return true
}

func (vm *VM) isWordBoundary(pos int) bool {
	// Check if we're at a transition between word and non-word characters
	prevChar, _ := vm.input.Context(pos)
	currChar, _ := vm.input.Step(pos)

	prevIsWord := isWordChar(prevChar)
	currIsWord := isWordChar(currChar)

	// Boundary exists when exactly one is a word char
	return prevIsWord != currIsWord
}

func isWordChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}
