%axiom ExpressionList
%preamble ParserPreallocMem

%%

ExpressionList : Expression NEWLINE ExpressionList
{
    fmt.Println("Expression Value:", *$1.Value.(*int64))
} | Expression NEWLINE
{
    fmt.Println("Expression Value:", *$1.Value.(*int64))
};

Expression : LPAREN Args RPAREN
{
    $$.Value = $2.Value;
} | NUMBER
{
    $$.Value = $1.Value;
};
   
Args : Expression SPACE Expression SPACE BinaryOp
{
    var first_op = *$1.Value.(*int64)
    var second_op = *$3.Value.(*int64)
    var operator = *$5.Value.(*string)

    value, err := binaryOp(first_op, second_op, operator) 
    if (err != nil) {
        fmt.Errorf("could not apply binary opeator")
    }
    $$.Value = &value
};

BinaryOp : PLUS
{
    var op = opParserPools[thread].Get()
    *op = "+"
    $$.Value = op
} | MINUS
{
    var op = opParserPools[thread].Get()
    *op = "-"
    $$.Value = op;
};

%%

import (
	"math"
    "errors"
)

var parserPools []*gopapageno.Pool[int64]
var opParserPools []*gopapageno.Pool[string]

func binaryOp(first_op, second_op int64, operator string) (int64, error) {
    switch (operator) {
        case "+":
            return first_op + second_op, nil
        case "-":
            return first_op - second_op, nil
        default:
            return 0, errors.New("binary operator unhandled")
    }
}

func ParserPreallocMem(inputSize int, numThreads int) {
	opParserPools = make([]*gopapageno.Pool[string], numThreads)
	parserPools = make([]*gopapageno.Pool[int64], numThreads)

	avgCharsPerNumber := float64(2)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		parserPools[i] = gopapageno.NewPool[int64](poolSizePerThread)
		opParserPools[i] = gopapageno.NewPool[string](poolSizePerThread)
	}
}
