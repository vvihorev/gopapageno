// Code generated by Gopapageno; DO NOT EDIT.
package main

import (
	"github.com/giornetta/gopapageno"
	"strings"
	"fmt"
)

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


// Non-terminals
const (
	Args = gopapageno.TokenEmpty + 1 + iota
	BinaryOp
	Expression
	ExpressionList
)

// Terminals
const (
	LPAREN = gopapageno.TokenTerm + 1 + iota
	MINUS
	NEWLINE
	NUMBER
	PLUS
	RPAREN
	SPACE
)

func SprintToken[ValueType any](root *gopapageno.Token) string {
	var sprintRec func(t *gopapageno.Token, sb *strings.Builder, indent string)

	sprintRec = func(t *gopapageno.Token, sb *strings.Builder, indent string) {
		if t == nil {
			return
		}

		sb.WriteString(indent)
		if t.Next == nil {
			sb.WriteString("└── ")
			indent += "    "
		} else {
			sb.WriteString("├── ")
			indent += "|   "
		}

		switch t.Type {
		case Args:
			sb.WriteString("Args")
		case BinaryOp:
			sb.WriteString("BinaryOp")
		case Expression:
			sb.WriteString("Expression")
		case ExpressionList:
			sb.WriteString("ExpressionList")
		case gopapageno.TokenEmpty:
			sb.WriteString("Empty")
		case LPAREN:
			sb.WriteString("LPAREN")
		case MINUS:
			sb.WriteString("MINUS")
		case NEWLINE:
			sb.WriteString("NEWLINE")
		case NUMBER:
			sb.WriteString("NUMBER")
		case PLUS:
			sb.WriteString("PLUS")
		case RPAREN:
			sb.WriteString("RPAREN")
		case SPACE:
			sb.WriteString("SPACE")
		case gopapageno.TokenTerm:
			sb.WriteString("Term")
		default:
			sb.WriteString("Unknown")
		}

		if t.Value != nil {
	        switch v := t.Value.(type) {
	        case *int64:
	            sb.WriteString(fmt.Sprintf(": %v", *v))
	        case *string:
	            sb.WriteString(fmt.Sprintf(": %v", *v))
	        }

			if v, ok := any(t.Value).(*ValueType); ok {
				sb.WriteString(fmt.Sprintf(": %v", *v))
			}
		}
		
		sb.WriteString("\n")
		
		sprintRec(t.Child, sb, indent)
		sprintRec(t.Next, sb, indent[:len(indent)-4])
	}

	var sb strings.Builder
	
	sprintRec(root, &sb, "")
	
	return sb.String()
}

func NewGrammar() *gopapageno.Grammar {
	numTerminals := uint16(8)
	numNonTerminals := uint16(5)

	maxRHSLen := 5
	rules := []gopapageno.Rule{
		{ExpressionList, []gopapageno.TokenType{Expression, NEWLINE}, gopapageno.RuleSimple},
		{ExpressionList, []gopapageno.TokenType{Expression, NEWLINE, ExpressionList}, gopapageno.RuleSimple},
		{Args, []gopapageno.TokenType{Expression, SPACE, Expression, SPACE, BinaryOp}, gopapageno.RuleSimple},
		{Expression, []gopapageno.TokenType{LPAREN, Args, RPAREN}, gopapageno.RuleSimple},
		{BinaryOp, []gopapageno.TokenType{MINUS}, gopapageno.RuleSimple},
		{Expression, []gopapageno.TokenType{NUMBER}, gopapageno.RuleSimple},
		{BinaryOp, []gopapageno.TokenType{PLUS}, gopapageno.RuleSimple},
	}
	compressedRules := []uint16{0, 0, 5, 3, 13, 32769, 46, 32770, 59, 32772, 62, 32773, 65, 0, 0, 2, 32771, 20, 32775, 28, 4, 0, 1, 4, 25, 4, 1, 0, 0, 0, 1, 3, 33, 0, 0, 1, 32775, 38, 0, 0, 1, 2, 43, 1, 2, 0, 0, 0, 1, 1, 51, 0, 0, 1, 32774, 56, 3, 3, 0, 2, 4, 0, 3, 5, 0, 2, 6, 0	}

	precMatrix := [][]gopapageno.Precedence{
		{gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecTakes, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecTakes},
		{gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecTakes, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecTakes},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecTakes, gopapageno.PrecEquals},
	}
	bitPackedMatrix := []uint64{
		91796036460631380, 2672464725262106754, 
	}

	fn := func(ruleDescription uint16, ruleFlags gopapageno.RuleFlags, lhs *gopapageno.Token, rhs []*gopapageno.Token, thread int){
		switch ruleDescription {
		case 0:
			ExpressionList0 := lhs
			Expression1 := rhs[0]
			NEWLINE2 := rhs[1]

			ExpressionList0.Child = Expression1
			Expression1.Next = NEWLINE2
			ExpressionList0.LastChild = NEWLINE2

			{
			    fmt.Println("Expression Value:", *Expression1.Value.(*int64))
			}
			_ = Expression1
			_ = NEWLINE2
		case 1:
			ExpressionList0 := lhs
			Expression1 := rhs[0]
			NEWLINE2 := rhs[1]
			ExpressionList3 := rhs[2]

			ExpressionList0.Child = Expression1
			Expression1.Next = NEWLINE2
			NEWLINE2.Next = ExpressionList3
			ExpressionList0.LastChild = ExpressionList3

			{
			    fmt.Println("Expression Value:", *Expression1.Value.(*int64))
			}
			_ = Expression1
			_ = NEWLINE2
			_ = ExpressionList3
		case 2:
			Args0 := lhs
			Expression1 := rhs[0]
			SPACE2 := rhs[1]
			Expression3 := rhs[2]
			SPACE4 := rhs[3]
			BinaryOp5 := rhs[4]

			Args0.Child = Expression1
			Expression1.Next = SPACE2
			SPACE2.Next = Expression3
			Expression3.Next = SPACE4
			SPACE4.Next = BinaryOp5
			Args0.LastChild = BinaryOp5

			{
			    var first_op = *Expression1.Value.(*int64)
			    var second_op = *Expression3.Value.(*int64)
			    var operator = *BinaryOp5.Value.(*string)
			
			    value, err := binaryOp(first_op, second_op, operator) 
			    if (err != nil) {
			        fmt.Errorf("could not apply binary opeator")
			    }
			    Args0.Value = &value
			}
			_ = Expression1
			_ = SPACE2
			_ = Expression3
			_ = SPACE4
			_ = BinaryOp5
		case 3:
			Expression0 := lhs
			LPAREN1 := rhs[0]
			Args2 := rhs[1]
			RPAREN3 := rhs[2]

			Expression0.Child = LPAREN1
			LPAREN1.Next = Args2
			Args2.Next = RPAREN3
			Expression0.LastChild = RPAREN3

			{
			    Expression0.Value = Args2.Value;
			}
			_ = LPAREN1
			_ = Args2
			_ = RPAREN3
		case 4:
			BinaryOp0 := lhs
			MINUS1 := rhs[0]

			BinaryOp0.Child = MINUS1
			BinaryOp0.LastChild = MINUS1

			{
			    var op = opParserPools[thread].Get()
			    *op = "-"
			    BinaryOp0.Value = op;
			}
			_ = MINUS1
		case 5:
			Expression0 := lhs
			NUMBER1 := rhs[0]

			Expression0.Child = NUMBER1
			Expression0.LastChild = NUMBER1

			{
			    Expression0.Value = NUMBER1.Value;
			}
			_ = NUMBER1
		case 6:
			BinaryOp0 := lhs
			PLUS1 := rhs[0]

			BinaryOp0.Child = PLUS1
			BinaryOp0.LastChild = PLUS1

			{
			    var op = opParserPools[thread].Get()
			    *op = "+"
			    BinaryOp0.Value = op
			}
			_ = PLUS1
		}
		_ = ruleFlags
	}

	return &gopapageno.Grammar{
		NumTerminals:  numTerminals,
		NumNonterminals: numNonTerminals,
		MaxRHSLength: maxRHSLen,
		Rules: rules,
		CompressedRules: compressedRules,
		PrecedenceMatrix: precMatrix,
		BitPackedPrecedenceMatrix: bitPackedMatrix,
		Func: fn,
		ParsingStrategy: gopapageno.AOPP,
		PreambleFunc: ParserPreallocMem,
	}
}

