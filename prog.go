package gore

import "fmt"

type OpCode int

const (
	OpMatch      OpCode = iota // Terminate success
	OpChar                     // Match specific rune
	OpCharClass                // Match char class
	OpAny                      // Match any (dot), usually valid utf8
	OpJmp                      // Jump to Offset
	OpSplit                    // Splits execution (try X, else Y)
	OpSave                     // Save position to capture register
	OpAssert                   // Zero-width assertion (Start/End line)
	OpLookaround               // Recursive check for lookaround
)

type Inst struct {
	Op         OpCode
	Val        rune          // For OpChar
	Ranges     []RuneRange   // For OpCharClass
	Negated    bool          // For OpCharClass
	Out        int           // Jump target 1 (primary)
	Out1       int           // Jump target 2 (alternative for Split)
	Idx        int           // Register index for OpSave
	Assert     AssertionType // For OpAssert
	Prog       *Prog         // For OpLookaround (sub-routine)
	LookNeg    bool          // Negative lookaround
	LookBehind bool          // Lookbehind
}

// Prog is a compiled regular expression program.
type Prog struct {
	Insts  []Inst
	Start  int // Entry point
	NumCap int // Number of capture registers needed
}

func (i Inst) String() string {
	switch i.Op {
	case OpMatch:
		return "match"
	case OpChar:
		return fmt.Sprintf("char %q", i.Val)
	case OpCharClass:
		neg := ""
		if i.Negated {
			neg = "^"
		}
		return fmt.Sprintf("class %s%v", neg, i.Ranges)
	case OpAny:
		return "any"
	case OpJmp:
		return fmt.Sprintf("jmp %d", i.Out)
	case OpSplit:
		return fmt.Sprintf("split %d, %d", i.Out, i.Out1)
	case OpSave:
		return fmt.Sprintf("save %d", i.Idx)
	case OpAssert:
		return fmt.Sprintf("assert %d", i.Assert)
	case OpLookaround:
		return fmt.Sprintf("look %v %d", i.LookNeg, i.Prog.Start)
	}
	return "?"
}
