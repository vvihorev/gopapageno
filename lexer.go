package gopapageno

import (
	"context"
	"errors"
	"fmt"
	"unsafe"
)

var (
	ErrInvalid = errors.New("invalid character")
)

type PreambleFunc func(sourceLen, concurrency int)

type LexerFunc func(rule int, text string, start int, end int, thread int, token *Token) LexResult

type Lexer struct {
	Automaton          LexerDFA
	CutPointsAutomaton LexerDFA
	Func               LexerFunc

	PreambleFunc PreambleFunc
}

type LexerDFAState struct {
	Transitions     [256]int
	IsFinal         bool
	AssociatedRules []int
}

type LexerDFA []LexerDFAState

// Scanner implements reading and tokenization.
type Scanner struct {
	Lexer *Lexer

	source      []byte
	cutPoints   []int
	concurrency int

	pools []*Pool[stack[Token]]
}

func (l *Lexer) Scanner(src []byte, opts *RunOptions) *Scanner {
	if opts.Concurrency < 1 {
		opts.Concurrency = 1
	}

	s := &Scanner{
		Lexer: l,

		source:      src,
		cutPoints:   []int{0},
		concurrency: 1,
	}

	s.cutPoints, s.concurrency = s.findCutPoints(opts.Concurrency)

	s.pools = make([]*Pool[stack[Token]], s.concurrency)

	if opts.AvgTokenLength < 1 {
		opts.AvgTokenLength = 1
	}

	stacksNum := stacksCount[Token](s.source, s.concurrency, opts.AvgTokenLength)

	for thread := 0; thread < s.concurrency; thread++ {
		s.pools[thread] = NewPool(stacksNum, WithConstructor(newStack[Token]))
	}

	return s
}

// findCutPoints cuts the source string at specific points determined by the lexer description file.
// It returns a slice containing the cut points indices in the source string, and the number of goroutines to spawn to handle them.
func (s *Scanner) findCutPoints(maxConcurrency int) ([]int, int) {
	sourceLen := len(s.source)
	avgBytesPerThread := sourceLen / maxConcurrency

	cutPoints := make([]int, maxConcurrency+1)
	cutPoints[0] = 0
	cutPoints[maxConcurrency] = len(s.source)

	for i := 1; i < maxConcurrency; i++ {
		startPos := cutPoints[i-1] + avgBytesPerThread

		pos := startPos
		state := s.Lexer.CutPointsAutomaton[0]

		for !state.IsFinal {
			if pos >= sourceLen {
				return append(cutPoints[0:i], cutPoints[maxConcurrency]), i
			}

			stateIdx := state.Transitions[s.source[pos]]

			//No more transitions are possible, reset the Automaton state
			if stateIdx == -1 {
				startPos = pos + 1
				state = s.Lexer.CutPointsAutomaton[0]
			} else {
				state = s.Lexer.Automaton[stateIdx]
			}
			pos++
		}
		cutPoints[i] = startPos
	}

	return cutPoints, maxConcurrency
}

func (s *Scanner) Lex(ctx context.Context) ([]*LOS[Token], error) {
	resultCh := make(chan lexResult, s.concurrency)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for thread := 0; thread < s.concurrency; thread++ {
		w := &scannerWorker{
			lexer:       s.Lexer,
			id:          thread,
			stackPool:   s.pools[thread],
			data:        s.source[s.cutPoints[thread]:s.cutPoints[thread+1]],
			pos:         0,
			startingPos: s.cutPoints[thread],
		}

		go w.lex(ctx, resultCh, errCh)
	}

	lexResults := make([]*LOS[Token], s.concurrency)
	completed := 0

	for completed < s.concurrency {
		select {
		case result := <-resultCh:
			lexResults[result.threadID] = result.tokens
			completed++
		case err := <-errCh:
			cancel()
			return nil, err
		}
	}

	return lexResults, nil
}

// worker implements the tokenizing logic on a subset of the source string.
type scannerWorker struct {
	lexer *Lexer

	id        int
	stackPool *Pool[stack[Token]]

	data []byte
	pos  int

	startingPos int
}

type lexResult struct {
	threadID int
	tokens   *LOS[Token]
}

// lex is the lexing function executed in parallel by each thread.
func (w *scannerWorker) lex(ctx context.Context, resultCh chan<- lexResult, errCh chan<- error) {
	los := NewLOS[Token](w.stackPool)

	var token Token

	for {
		token.Value = nil
		result := w.next(&token)
		if result != LexOK {
			if result == LexEOF {
				resultCh <- lexResult{
					threadID: w.id,
					tokens:   los,
				}
				return
			}

			errCh <- ErrInvalid
			return
		}

		los.Push(token)
	}
}

type LexResult uint8

const (
	LexOK LexResult = iota
	LexSkip
	LexErr
	LexEOF
)

// next scans the input text and returns the next Token.
func (w *scannerWorker) next(token *Token) LexResult {
	for {
		var lastFinalStateReached *LexerDFAState = nil
		var lastFinalStatePos int

		startPos := w.pos
		state := &w.lexer.Automaton[0]
		for {
			// If we reach the end of the source data, return EOF.
			if w.pos == len(w.data) {
				return LexEOF
			}

			stateIdx := state.Transitions[w.data[w.pos]]

			// If we are in an invalid state:
			if stateIdx == -1 {
				// If we haven't reached any final state so far, return an error.
				if lastFinalStateReached == nil {
					fmt.Printf("could not parse token '%s' in surrounding %s at position %d\n", string(w.data[startPos:w.pos+1]), string(w.data), startPos)
					return LexErr
				}

				result := w.advance(token, lastFinalStatePos, lastFinalStateReached, startPos)
				if result == LexSkip {
					break
				}

				return result
			}

			state = &w.lexer.Automaton[stateIdx]

			// If the state is not final, keep lexing.
			if !state.IsFinal {
				w.pos++
				continue
			}

			lastFinalStateReached = state
			lastFinalStatePos = w.pos

			if w.pos == len(w.data)-1 {
				result := w.advance(token, lastFinalStatePos, lastFinalStateReached, startPos)
				if result == LexSkip {
					break
				}

				return result
			}

			w.pos++
		}
	}
}

func (w *scannerWorker) advance(token *Token, lastFinalStatePos int, lastFinalStateReached *LexerDFAState, startPos int) LexResult {
	w.pos = lastFinalStatePos + 1
	ruleNum := lastFinalStateReached.AssociatedRules[0]

	// TODO: should be changed to safe code when Run supports no-op []byte to string conversion
	//text := unsafe.String(unsafe.SliceData(w.data[startPos:w.pos]), w.pos - startPos)
	textBytes := w.data[startPos:w.pos]
	text := *(*string)(unsafe.Pointer(&textBytes))

	// Compute absolute start & end position of the current token in the source file.
	tokenStart := w.startingPos + startPos
	tokenEnd := tokenStart + w.pos - startPos - 1

	if text == "\n" {
		token.Text = "\\n"
	} else {
		token.Text = text
	}

	return w.lexer.Func(ruleNum, text, tokenStart, tokenEnd, w.id, token)
}
