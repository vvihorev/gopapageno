package gopapageno

/*
type Symbol struct {
	Token      uint16
	precedence precedence
	Value      interface{}
	Next       *Symbol
	Child      *Symbol
}
*/

type Token struct {
	Type  TokenType
	Value any
	// Lexeme string

	Precedence Precedence

	Next  *Token
	Child *Token
}

type TokenType uint16

const (
	TokenEmpty TokenType = 0
	TokenTerm  TokenType = 0x8000
)

func (t TokenType) IsTerminal() bool {
	return t >= 0x8000
}

func (t TokenType) Value() uint16 {
	return uint16(0x7FFF & t)
}
