%cut \n

LPAR    \(
RPAR    \)
PLUS    \+
TIMES   \*
DIGIT   [0-9]
SPACE   [ \t]
NEWLINE [\r\n]

%%

{LPAR}
{
	*genSym = gopapageno.Symbol{LPAR, 0, nil, nil, nil}
	return _LEX_CORRECT
}
{RPAR}
{
	*genSym = gopapageno.Symbol{RPAR, 0, nil, nil, nil}
	return _LEX_CORRECT
}
{TIMES}
{
	*genSym = gopapageno.Symbol{TIMES, 0, nil, nil, nil}
	return _LEX_CORRECT
}
{PLUS}
{
	*genSym = gopapageno.Symbol{PLUS, 0, nil, nil, nil}
	return _LEX_CORRECT
}
{DIGIT}+
{
	num := lexerInt64Pools[thread].Get()

	*num, err := strconv.ParseInt(yytext, 10, 64)
	if err != nil {
		return _ERROR
	}

	*genSym = gopapageno.Symbol{NUMBER, 0, num, nil, nil}
	return _LEX_CORRECT
}
{SPACE}
{
	return _SKIP
}
{NEWLINE}
{
	return _SKIP
}
.
{
	return _ERROR
}

%%
import (
	"math"
	"strconv"
)

var lexerInt64Pools []*gopapageno.Pool[int64]

// lexerPreallocMem initializes all the memory pools required by the lexer.
func lexerPreallocMem(inputSize int, numThreads int) {
	lexerInt64Pools = make([]*gopapageno.Pool[int64], numThreads)
	
	avgCharsPerNumber := float64(4)
	
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		lexerInt64Pools[i] = gopapageno.NewPool[int64](poolSizePerThread)
	}
}