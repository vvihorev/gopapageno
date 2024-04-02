package gopapageno

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"runtime"
	"time"
)

/*
parsingStats contains some statistics about the parse.
*/
type parsingStats struct {
	NumLexThreads                           int
	NumParseThreads                         int
	StackPoolSizes                          []int
	StackPoolNewNonterminalsSizes           []int
	StackPtrPoolSizes                       []int
	StackPoolSizeFinalPass                  int
	StackPoolNewNonterminalsSizeFinalPass   int
	StackPtrPoolSizeFinalPass               int
	AllocMemTime                            time.Duration
	CutPoints                               []int
	LexTimes                                []time.Duration
	LexTimeTotal                            time.Duration
	NumTokens                               []int
	NumTokensTotal                          int
	ParseTimes                              []time.Duration
	RecombiningStacksTime                   time.Duration
	ParseTimeFinalPass                      time.Duration
	ParseTimeTotal                          time.Duration
	RemainingStacks                         []int
	RemainingStacksNewNonterminals          []int
	RemainingStackPtrs                      []int
	RemainingStacksFinalPass                int
	RemainingStacksNewNonterminalsFinalPass int
	RemainingStackPtrsFinalPass             int
}

// Stats contains some statistics that may be checked after a call to ParseString or ParseFile
// var Stats parsingStats

type Rule struct {
	Lhs TokenType
	Rhs []TokenType
}

type Parser struct {
	Lexer       *Lexer
	concurrency int

	NumTerminals    uint16
	NumNonterminals uint16

	MaxRHSLength    int
	Rules           []Rule
	CompressedRules []uint16

	PrecedenceMatrix          [][]Precedence
	BitPackedPrecedenceMatrix []uint64

	Func func(rule uint16, lhs *Token, rhs []*Token, thread int)

	logger *log.Logger
}

func (p *Parser) Concurrency() int {
	return p.concurrency
}

type ParserOpt func(p *Parser)

func WithConcurrency(n int) ParserOpt {
	return func(p *Parser) {
		if n <= 0 {
			n = 1
		}

		p.concurrency = n
	}
}

func WithLogging(logger *log.Logger) ParserOpt {
	return func(p *Parser) {
		p.logger = logger
	}
}

func (p *Parser) Parse(ctx context.Context, src []byte) (*Token, error) {
	// Instantiate no-op logger if it's not set.
	if p.logger == nil {
		p.logger = log.New(io.Discard, "", 0)
	}

	if p.concurrency <= 0 {
		p.concurrency = 1
	}

	scanner := p.Lexer.Scanner(src, ScannerWithConcurrency(p.concurrency))

	srcLen := len(src)
	avgCharsPerToken := 12.5

	stackPoolBaseSize := math.Ceil(float64(srcLen) / avgCharsPerToken / float64(stackSize) / float64(p.concurrency))
	stackPtrPoolBaseSize := math.Ceil(float64(srcLen) / avgCharsPerToken / float64(stackPtrSize) / float64(p.concurrency))

	pools := make([]*Pool[stack[Token]], p.concurrency)
	ptrPools := make([]*Pool[stackPtr], p.concurrency)

	for thread := 0; thread < p.concurrency; thread++ {
		pools[thread] = NewPool[stack[Token]](int(stackPoolBaseSize * 0.8))
		ptrPools[thread] = NewPool[stackPtr](int(stackPtrPoolBaseSize))
	}

	stackPoolFinalPass := NewPool[stack[Token]](int(math.Ceil(stackPoolBaseSize * 0.1 * float64(p.concurrency))))
	stackPoolNewNonterminalsFinalPass := NewPool[stack[Token]](int(math.Ceil(stackPoolBaseSize * 0.05 * float64(p.concurrency))))
	stackPtrPoolFinalPass := NewPool[stackPtr](int(math.Ceil(stackPtrPoolBaseSize * 0.1)))

	runtime.GC()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tokens, err := scanner.Lex(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not lex: %w", err)
	}
	// If there are not enough stacks in the input, reduce the number of threads.
	// The input is split by splitting stacks, not stack contents.
	if tokens.NumStacks() < p.concurrency {
		// TODO: Move this somewhere else?
		// TODO: Log?
		p.concurrency = tokens.NumStacks()
	}

	tokensLists, err := tokens.Split(p.concurrency)
	if err != nil {
		return nil, err
	}

	resultCh := make(chan parseResult)
	errCh := make(chan error, 1)

	workers := make([]*parserWorker, p.concurrency)
	for thread := 0; thread < p.concurrency; thread++ {
		var nextToken *Token

		if thread < p.concurrency-1 {
			nextInputListIter := tokensLists[thread+1].HeadIterator()
			nextToken = nextInputListIter.Next()
		}

		workers[thread] = &parserWorker{
			parser: p,
			id:     thread,

			stackPool:    pools[thread],
			ptrStackPool: ptrPools[thread],
		}

		go workers[thread].parse(ctx,
			tokensLists[thread],
			nextToken,
			false,
			resultCh,
			errCh)
	}

	parseResults := make([]parseResult, p.concurrency)
	completed := 0

	for completed < p.concurrency {
		select {
		case result := <-resultCh:
			parseResults[result.threadNum] = result
			completed++
		case err := <-errCh:
			cancel()
			return nil, err
		}
	}

	//If the number of threads is greater than one, a final pass is required
	if p.concurrency > 1 {
		//Create the final input by joining together the stacks from the previous step
		finalPassInput := NewLOS[Token](stackPoolFinalPass)

		for i := 0; i < p.concurrency; i++ {
			iterator := parseResults[i].stack.HeadIterator()

			//Ignore the first token
			iterator.Next()

			sym := iterator.Next()
			for sym != nil {
				finalPassInput.Push(*sym)
				sym = iterator.Next()
			}
		}

		workers[0].stackPool = stackPoolNewNonterminalsFinalPass
		workers[0].ptrStackPool = stackPtrPoolFinalPass

		p.concurrency = 1

		go workers[0].parse(ctx,
			finalPassInput,
			nil,
			true,
			resultCh,
			errCh)

		select {
		case result := <-resultCh:
			parseResults[0] = result
		case err := <-errCh:
			cancel()
			return nil, err
		}
	}

	//Pop tokens from the stack until a nonterminal is found
	sym := parseResults[0].stack.Pop()

	for sym.Type.IsTerminal() {
		sym = parseResults[0].stack.Pop()
	}

	return sym, nil
}

type parserWorker struct {
	parser *Parser

	id int

	stackPool    *Pool[stack[Token]]
	ptrStackPool *Pool[stackPtr]
}

type parseResult struct {
	threadNum int
	stack     *listOfStackPtrs
}

func (w *parserWorker) parse(ctx context.Context, tokens *LOS[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

	newNonTerminalsList := NewLOS[Token](w.stackPool)
	stack := newLosPtr(w.ptrStackPool)

	// If the thread is the first, push a # onto the stack
	// Otherwise, push the first inputToken onto the stack
	if w.id == 0 {
		stack.Push(&Token{
			Type:  TokenTerm,
			Value: nil,
			// Lexeme:     "",
			Precedence: PrecEmpty,
			Next:       nil,
			Child:      nil,
		})
	} else {
		t := tokensIt.Next()
		t.Precedence = PrecEmpty

		stack.Push(t)
	}

	// If the thread is the last, push a # onto the tokens m
	// Otherwise, push onto the tokens m the first inputToken of the next tokens m
	if w.id == w.parser.concurrency-1 {
		tokens.Push(Token{
			Type:  TokenTerm,
			Value: nil,
			// Lexeme:     "",
			Precedence: PrecEmpty,
			Next:       nil,
			Child:      nil,
		})
	} else {
		tokens.Push(*nextToken)
	}

	numYieldsPrec := 0

	var pos int
	var lhsToken *Token

	var rhs []TokenType
	var rhsTokens []*Token

	rhsBuf := make([]TokenType, w.parser.MaxRHSLength)
	rhsTokensBuf := make([]*Token, w.parser.MaxRHSLength)

	newNonTerm := &Token{
		Type:  TokenEmpty,
		Value: nil,
		// Lexeme:     "",
		Precedence: PrecEmpty,
		Next:       nil,
		Child:      nil,
	}

	i := 0
	// Iterate over the tokens m
	for inputToken := tokensIt.Next(); inputToken != nil; i++ {
		//If the current inputToken is a non-terminal, push it onto the stack with no precedence relation
		if !inputToken.Type.IsTerminal() {
			inputToken.Precedence = PrecEmpty
			stack.Push(inputToken)

			inputToken = tokensIt.Next()
			continue
		}

		//Find the first terminal on the stack and get the precedence between it and the current tokens inputToken
		firstTerminal := stack.FirstTerminal()
		prec := w.parser.precedence(firstTerminal.Type, inputToken.Type)
		switch prec {
		// If it yields precedence, push the tokens inputToken onto the stack with that precedence relation.
		// Also increment the counter of the number of tokens yielding precedence.
		case PrecYields:
			inputToken.Precedence = PrecYields
			stack.Push(inputToken)
			numYieldsPrec++

			inputToken = tokensIt.Next()
		// If it's equal in precedence, push the tokens inputToken onto the stack with that precedence relation
		case PrecEquals:
			inputToken.Precedence = PrecEquals
			stack.Push(inputToken)

			inputToken = tokensIt.Next()
		// If it takes precedence, the next action depends on whether there are tokens that yield precedence onto the stack.
		case PrecTakes:
			//If there are no tokens yielding precedence on the stack, push the tokens inputToken onto the stack with take precedence as precedence relation
			//Otherwise, perform a reduction
			if numYieldsPrec == 0 {
				inputToken.Precedence = PrecTakes
				stack.Push(inputToken)

				inputToken = tokensIt.Next()
			} else {
				pos = w.parser.MaxRHSLength - 1

				var token *Token
				// Pop tokens from the stack until one that yields precedence is reached, saving them in rhsBuf
				for token = stack.Pop(); token.Precedence != PrecYields; token = stack.Pop() {
					rhsTokensBuf[pos] = token
					rhsBuf[pos] = token.Type
					pos--
				}
				rhsTokensBuf[pos] = token
				rhsBuf[pos] = token.Type

				//Pop one last token, if it's a non-terminal add it to rhsBuf, otherwise ignore it (push it again onto the stack)
				token = stack.Pop()
				if token.Type.IsTerminal() {
					stack.Push(token)
				} else {
					pos--
					rhsTokensBuf[pos] = token
					rhsBuf[pos] = token.Type

					stack.UpdateFirstTerminal()
				}

				//Obtain the actual rhs from the buffers
				rhsTokens = rhsTokensBuf[pos:]
				rhs = rhsBuf[pos:]

				//Find corresponding lhs and ruleNum
				lhs, ruleNum := w.parser.findMatch(rhs)
				if lhs == TokenEmpty {
					errCh <- fmt.Errorf("could not find match for rhs %v", rhs)
					return
				}

				//Push the new nonterminal onto the appropriate m to save it
				newNonTerm.Type = lhs
				lhsToken = newNonTerminalsList.Push(*newNonTerm)

				//Execute the semantic action
				w.parser.Func(ruleNum, lhsToken, rhsTokens, w.id)

				//Push the new nonterminal onto the stack
				stack.Push(lhsToken)

				//Decrement the counter of the number of tokens yielding precedence
				numYieldsPrec--
			}
		//If there's no precedence relation, abort the parsing
		default:
			errCh <- fmt.Errorf("no precedence relation found")
			return
		}
	}

	resultCh <- parseResult{w.id, &stack}
}

func (p *Parser) precedence(t1 TokenType, t2 TokenType) Precedence {
	v1 := t1.Value()
	v2 := t2.Value()

	flatElementPos := v1*p.NumTerminals + v2
	elem := p.BitPackedPrecedenceMatrix[flatElementPos/32]
	pos := uint((flatElementPos % 32) * 2)

	return Precedence((elem >> pos) & 0x3)
}

func (p *Parser) findMatch(rhs []TokenType) (TokenType, uint16) {
	var pos uint16

	for _, key := range rhs {
		//Skip the value and rule num for each node (except the last)
		pos += 2
		numIndices := p.CompressedRules[pos]
		if numIndices == 0 {
			return TokenEmpty, 0
		}

		pos++
		low := uint16(0)
		high := numIndices - 1
		startPos := pos
		foundNext := false

		for low <= high {
			indexPos := low + (high-low)/2
			pos = startPos + indexPos*2
			curKey := p.CompressedRules[pos]

			if uint16(key) < curKey {
				high = indexPos - 1
			} else if uint16(key) > curKey {
				low = indexPos + 1
			} else {
				pos = p.CompressedRules[pos+1]
				foundNext = true
				break
			}
		}
		if !foundNext {
			return TokenEmpty, 0
		}
	}

	return TokenType(p.CompressedRules[pos]), p.CompressedRules[pos+1]
}
