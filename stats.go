package gopapageno

import "time"

// Stats contains some useful data about the parsing process.
type Stats struct {
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
