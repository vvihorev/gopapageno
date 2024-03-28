package gopapageno

import (
	"context"
	"errors"
	"math"
	"unsafe"
)

var (
	ErrEOF         = errors.New("reached end of file")
	ErrNotAccepted = errors.New("no final state reached")
	ErrSkip        = errors.New("skip current character")
	ErrInvalid     = errors.New("invalid character")
)

type LexerFunc func(rule int, text string, token *Token) error

type Lexer struct {
	Automaton          LexerDFA
	CutPointsAutomaton LexerDFA
	Func               LexerFunc
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

	source []byte

	cutPoints []int

	concurrency int
}

type ScannerOpt func(*Scanner)

func (l *Lexer) Scanner(src []byte, opts ...ScannerOpt) *Scanner {
	s := &Scanner{
		Lexer: l,

		source:      src,
		cutPoints:   []int{0},
		concurrency: 1,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithConcurrency accepts a desired number of goroutines to spawn during lexical analysis.
// It will look for suitable cut points in the source string and set the actual concurrency level accordingly.
func ScannerWithConcurrency(n int) ScannerOpt {
	return func(s *Scanner) {
		if n <= 0 {
			n = 1
		}

		// TODO: Log if result < n?
		s.cutPoints, s.concurrency = s.findCutPoints(n)
	}
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

func (s *Scanner) Lex(ctx context.Context) (*LOS[Token], error) {
	resultCh := make(chan lexResult)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// TODO: Should this be moved?
	sourceLen := len(s.source)
	avgCharsPerToken := 12.5

	stackPoolBaseSize := math.Ceil(float64(sourceLen) / avgCharsPerToken / float64(stackSize) / float64(s.concurrency))

	for thread := 0; thread < s.concurrency; thread++ {
		w := &scannerWorker{
			lexer:     s.Lexer,
			id:        thread,
			stackPool: NewPool[stack[Token]](int(stackPoolBaseSize * 1.2)),
			data:      s.source[s.cutPoints[thread]:s.cutPoints[thread+1]],
			pos:       0,
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

	// Merge results of different goroutines,
	// guaranteeing ordering amongst them.
	for i := 1; i < s.concurrency; i++ {
		lexResults[0].Merge(lexResults[i])
	}

	return lexResults[0], nil
}

// worker implements the tokenizing logic on a subset of the source string.
type scannerWorker struct {
	lexer *Lexer

	id        int
	stackPool *Pool[stack[Token]]

	data []byte
	pos  int
}

type lexResult struct {
	threadID int
	tokens   *LOS[Token]
}

// lex is the lexing function executed in parallel by each thread.
func (w *scannerWorker) lex(ctx context.Context, resultCh chan<- lexResult, errCh chan<- error) {
	los := NewLOS[Token](w.stackPool)

	for {
		token, err := w.next()
		if err != nil {
			if errors.Is(err, ErrEOF) {
				resultCh <- lexResult{
					threadID: w.id,
					tokens:   los,
				}
				return
			}

			errCh <- err
			return
		}

		los.Push(token)
	}
}

// next scans the input text and returns the next Token.
func (w *scannerWorker) next() (Token, error) {
	var lastFinalStateReached *LexerDFAState = nil
	var lastFinalStatePos int

	startPos := w.pos
	state := w.lexer.Automaton[0]

	for {
		// If we reach the end of the source data, return EOF.
		if w.pos == len(w.data) {
			return Token{}, ErrEOF
		}

		stateIdx := state.Transitions[w.data[w.pos]]

		// If we are in an invalid state:
		if stateIdx == -1 {
			// If we haven't reached any final state so far, return an error.
			if lastFinalStateReached == nil {
				// TODO: This returned _ERROR, which error is it?
				return Token{}, ErrNotAccepted
			}

			// TODO: Duplicated code
			// Otherwise, return it.
			w.pos = lastFinalStatePos + 1
			ruleNum := lastFinalStateReached.AssociatedRules[0]

			//TODO should be changed to safe code when Go supports no-op []byte to string conversion
			textBytes := w.data[startPos:w.pos]
			text := *(*string)(unsafe.Pointer(&textBytes))

			var token Token
			err := w.lexer.Func(ruleNum, text, &token)
			if err != nil {
				if errors.Is(err, ErrSkip) {
					w.pos++
					continue
				}

				return Token{}, err
			}

			return token, nil
		}

		state = w.lexer.Automaton[stateIdx]

		// If the state is not final, keep lexing.
		if !state.IsFinal {
			w.pos++
			continue
		}

		lastFinalStateReached = &state
		lastFinalStatePos = w.pos

		if w.pos == len(w.data)-1 {
			w.pos = lastFinalStatePos + 1
			ruleNum := lastFinalStateReached.AssociatedRules[0]

			//TODO should be changed to safe code when Go supports no-op []byte to string conversion
			textBytes := w.data[startPos:w.pos]
			text := *(*string)(unsafe.Pointer(&textBytes))

			var token Token
			err := w.lexer.Func(ruleNum, text, &token)
			if err != nil {
				if errors.Is(err, ErrSkip) {
					w.pos++
					continue
				}

				return Token{}, err
			}

			return token, nil
		}
	}
}
