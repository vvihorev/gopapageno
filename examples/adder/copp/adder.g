%axiom S
%preamble ParserPreallocMem

%%

S : E
{
    $$.Value = $1.Value
};

E : (T PLUS)+ T
{
    var firstValue, secondValue int64

    if !ruleFlags.Has(gopapageno.RuleAppend) {
        $$.Value = parserPools[thread].Get()

        firstValue = *$1.Value.(*int64)
    } else {
        firstValue = *$$.Value.(*int64)
    }

    secondValue = *$3.Value.(*int64)
    *$$.Value.(*int64) = firstValue + secondValue
};

T : LPAR T RPAR
{
    $$.Value = $2.Value
} | LPAR E RPAR
{
  $$.Value = $2.Value
} | NUMBER
{
    $$.Value = $1.Value
};

%%

import (
	"math"
)

var parserPools []*gopapageno.Pool[int64]

func ParserPreallocMem(inputSize int, numThreads int) {
	parserPools = make([]*gopapageno.Pool[int64], numThreads)

	avgCharsPerNumber := float64(2)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		parserPools[i] = gopapageno.NewPool[int64](poolSizePerThread)
	}
}