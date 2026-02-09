package gore

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenChar         // Literal character
	TokenDot          // .
	TokenPipe         // |
	TokenLParen       // (
	TokenRParen       // )
	TokenLBracket     // [
	TokenRBracket     // ]
	TokenPlus         // +
	TokenStar         // *
	TokenQuestion     // ?
	TokenCarret       // ^
	TokenDollar       // $
	TokenBackslash    // \
	TokenLBrace       // {
	TokenRBrace       // }
	TokenQuestP       // (? for extension
	TokenQuestPEqb    // (?=
	TokenQuestPExcl   // (?!
	TokenQuestPLTEqb  // (?<=
	TokenQuestPLTExcl // (?<!
	TokenQuestPColon  // (?:
	TokenQuestPName   // (?P<
)

type Token struct {
	Type  TokenType
	Val   rune
	Start int
	End   int
}
