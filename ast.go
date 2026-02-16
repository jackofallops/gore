package gore

// NodeType identifies the type of AST node.
type NodeType int

const (
	NodeLiteral NodeType = iota
	NodeConcat
	NodeAlternate
	NodeQuantifier
	NodeCapture
	NodeAssertion
	NodeLookaround
	NodeCharClass   // [new]
	NodeBackreference
)

// Node is the base interface for AST nodes.
type Node interface {
	Type() NodeType
}

// Literal matches a sequence of runes.
type Literal struct {
	Runes    []rune
	FoldCase bool // Case-insensitive matching
}

func (n *Literal) Type() NodeType { return NodeLiteral }

// Concat matches a sequence of nodes.
type Concat struct {
	Nodes []Node
}

func (n *Concat) Type() NodeType { return NodeConcat }

// Alternate matches one of several branches.
type Alternate struct {
	Nodes []Node
}

func (n *Alternate) Type() NodeType { return NodeAlternate }

// Quantifier matches a node repeated min..max times.
type Quantifier struct {
	Body   Node
	Min    int
	Max    int // -1 for infinity
	Greedy bool
}

func (n *Quantifier) Type() NodeType { return NodeQuantifier }

// Capture creates a capture group.
type Capture struct {
	Body  Node
	Index int    // 1-based index
	Name  string // Optional name
}

func (n *Capture) Type() NodeType { return NodeCapture }

// Assertion matches a position without consuming characters.
type AssertionType int

const (
	AssertStartText       AssertionType = iota // ^
	AssertEndText                              // $
	AssertWordBoundary                         // \b
	AssertNotWordBoundary                      // \B
	AssertStringStart                          // \A
	AssertStringEnd                            // \Z
	AssertAbsoluteEnd                          // \z
)

type Assertion struct {
	Kind      AssertionType
	Multiline bool // True if ^ or $ should behave in multiline mode
}

func (n *Assertion) Type() NodeType { return NodeAssertion }

// Lookaround is a zero-width assertion that matches a pattern.
type Lookaround struct {
	Body     Node
	Negative bool // True for (?!...) and (?<!...)
	Behind   bool // True for (?<=...) and (?<!...)
}

func (n *Lookaround) Type() NodeType { return NodeLookaround }

// CharClass represents [a-z0-9] or [^a-z].
type CharClass struct {
	Ranges   []RuneRange
	Negated  bool
	FoldCase bool // Case-insensitive matching
}

type RuneRange struct {
	Lo, Hi rune
}

func (n *CharClass) Type() NodeType { return NodeCharClass }

// Backreference refers to a previously captured group.
type Backreference struct {
	Index int // 1-based index of the capture group
}

func (n *Backreference) Type() NodeType { return NodeBackreference }
