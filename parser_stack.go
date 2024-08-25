package gopapageno

import (
	"fmt"
)

// parserStack is the base data structure used by workers during parsing.
type parserStack struct {
	*LOPS[Token]

	firstTerminalStack *stack[*Token]
	firstTerminalPos   int
}

// newParserStack creates an empty parserStack.
func newParserStack(pool *Pool[stack[*Token]]) *parserStack {
	return &parserStack{
		LOPS: NewLOPS[Token](pool),
	}
}

// Push pushes a token pointer in the ParserStack.
// It returns the pointer itself.
func (s *parserStack) Push(token *Token) *Token {
	s.LOPS.Push(token)

	//If the token is a terminal update the firstTerminal pointer
	if token.Type.IsTerminal() {
		s.firstTerminalStack = s.LOPS.cur
		s.firstTerminalPos = s.LOPS.cur.Tos - 1
	}

	return token
}

// FirstTerminal returns a pointer to the first terminal token on the stack.
func (s *parserStack) FirstTerminal() *Token {
	return s.firstTerminalStack.Data[s.firstTerminalPos]
}

// UpdateFirstTerminal should be used after a reduction in order to update the first terminal counter.
// In fact, in order to save some time, only the Push operation automatically updates the first terminal pointer,
// while the Pop operation does not.
func (s *parserStack) UpdateFirstTerminal() {
	s.firstTerminalStack, s.firstTerminalPos = s.findFirstTerminal()
}

// findFirstTerminal computes the first terminal on the stacks.
// This function is for internal usage only.
func (s *parserStack) findFirstTerminal() (*stack[*Token], int) {
	curStack := s.cur

	pos := curStack.Tos - 1

	for pos < s.headFirst {
		pos = -1
		if curStack.Prev == nil {
			return nil, 0
		}

		s.headFirst = 0

		curStack = curStack.Prev
		pos = curStack.Tos - 1
	}

	for !curStack.Data[pos].Type.IsTerminal() {
		pos--
		for pos < s.headFirst {
			pos = -1
			if curStack.Prev == nil {
				return nil, 0
			}

			s.headFirst = 0

			curStack = curStack.Prev
			pos = curStack.Tos - 1
		}
	}

	return curStack, pos
}

func (s *parserStack) LastNonterminal() (*Token, error) {
	for token := s.Pop(); token != nil; token = s.Pop() {
		if !token.Type.IsTerminal() {
			return token, nil
		}
	}

	return nil, fmt.Errorf("no nonterminal found")
}
