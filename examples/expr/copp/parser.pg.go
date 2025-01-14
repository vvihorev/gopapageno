// Code generated by Gopapageno; DO NOT EDIT.
package main

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"strings"
)

import (
	"math"
)

var parserInt64Pools []*gopapageno.Pool[int64]

// ParserPreallocMem initializes all the memory pools required by the semantic function of the parser.
func ParserPreallocMem(inputSize int, numThreads int) {
	parserInt64Pools = make([]*gopapageno.Pool[int64], numThreads)

	avgCharsPerNumber := float64(4)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		parserInt64Pools[i] = gopapageno.NewPool[int64](poolSizePerThread)
	}
}

// Non-terminals
const (
	D_E_P_T = gopapageno.TokenEmpty + 1 + iota
	D_T
	P
	S
)

// Terminals
const (
	DIVIDE = gopapageno.TokenTerm + 1 + iota
	LPAR
	NUMBER
	PLUS
	RPAR
)

func SprintToken[TokenValue any](root *gopapageno.Token) string {
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
		case D_E_P_T:
			sb.WriteString("D_E_P_T")
		case D_T:
			sb.WriteString("D_T")
		case P:
			sb.WriteString("P")
		case S:
			sb.WriteString("S")
		case gopapageno.TokenEmpty:
			sb.WriteString("Empty")
		case DIVIDE:
			sb.WriteString("DIVIDE")
		case LPAR:
			sb.WriteString("LPAR")
		case NUMBER:
			sb.WriteString("NUMBER")
		case PLUS:
			sb.WriteString("PLUS")
		case RPAR:
			sb.WriteString("RPAR")
		case gopapageno.TokenTerm:
			sb.WriteString("Term")
		default:
			sb.WriteString("Unknown")
		}
		if t.Value != nil {
			sb.WriteString(fmt.Sprintf(": %v", *t.Value.(*TokenValue)))
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
	numTerminals := uint16(6)
	numNonTerminals := uint16(5)

	maxRHSLen := 3
	rules := []gopapageno.Rule{
		{S, []gopapageno.TokenType{D_E_P_T}, gopapageno.RuleSimple},
		{D_T, []gopapageno.TokenType{D_E_P_T, DIVIDE, D_E_P_T}, gopapageno.RuleSimple},
		{P, []gopapageno.TokenType{D_E_P_T, PLUS, D_E_P_T}, gopapageno.RuleCyclic},
		{P, []gopapageno.TokenType{D_E_P_T, PLUS, D_T}, gopapageno.RuleCyclic},
		{P, []gopapageno.TokenType{D_E_P_T, PLUS, P}, gopapageno.RuleAppendLeft},
		{P, []gopapageno.TokenType{D_E_P_T, PLUS, P}, gopapageno.RuleAppendLeft},
		{S, []gopapageno.TokenType{D_T}, gopapageno.RuleSimple},
		{D_T, []gopapageno.TokenType{D_T, DIVIDE, D_E_P_T}, gopapageno.RuleSimple},
		{P, []gopapageno.TokenType{D_T, PLUS, D_E_P_T}, gopapageno.RuleCyclic},
		{P, []gopapageno.TokenType{D_T, PLUS, D_T}, gopapageno.RuleCyclic},
		{P, []gopapageno.TokenType{D_T, PLUS, P}, gopapageno.RuleAppendLeft},
		{P, []gopapageno.TokenType{D_T, PLUS, P}, gopapageno.RuleAppendLeft},
		{S, []gopapageno.TokenType{P}, gopapageno.RuleSimple},
		{P, []gopapageno.TokenType{P, PLUS, D_E_P_T}, gopapageno.RuleAppend},
		{P, []gopapageno.TokenType{P, PLUS, D_E_P_T}, gopapageno.RuleAppend},
		{P, []gopapageno.TokenType{P, PLUS, D_T}, gopapageno.RuleAppend},
		{P, []gopapageno.TokenType{P, PLUS, D_T}, gopapageno.RuleAppend},
		{P, []gopapageno.TokenType{P, PLUS, P}, gopapageno.RuleCombine},
		{P, []gopapageno.TokenType{P, PLUS, P}, gopapageno.RuleCombine},
		{P, []gopapageno.TokenType{P, PLUS, P}, gopapageno.RuleCombine},
		{P, []gopapageno.TokenType{P, PLUS, P}, gopapageno.RuleCombine},
		{S, []gopapageno.TokenType{S}, gopapageno.RuleSimple},
		{D_E_P_T, []gopapageno.TokenType{LPAR, S, RPAR}, gopapageno.RuleSimple},
		{D_E_P_T, []gopapageno.TokenType{NUMBER}, gopapageno.RuleSimple},
	}
	compressedRules := []uint16{0, 0, 6, 1, 15, 2, 48, 3, 81, 4, 104, 32770, 107, 32771, 120, 4, 0, 2, 32769, 22, 32772, 30, 0, 0, 1, 1, 27, 2, 1, 0, 0, 0, 3, 1, 39, 2, 42, 3, 45, 3, 2, 0, 3, 3, 0, 3, 5, 0, 4, 6, 2, 32769, 55, 32772, 63, 0, 0, 1, 1, 60, 2, 7, 0, 0, 0, 3, 1, 72, 2, 75, 3, 78, 3, 8, 0, 3, 9, 0, 3, 11, 0, 4, 12, 1, 32772, 86, 0, 0, 3, 1, 95, 2, 98, 3, 101, 3, 14, 0, 3, 16, 0, 3, 20, 0, 4, 21, 0, 0, 0, 1, 4, 112, 0, 0, 1, 32773, 117, 1, 22, 0, 1, 23, 0}

	maxPrefixLength := 5
	prefixes := [][]gopapageno.TokenType{
		{D_E_P_T, PLUS, D_E_P_T},
		{D_E_P_T, PLUS, D_T},
		{D_T, PLUS, D_E_P_T},
		{D_T, PLUS, D_T},
		{D_E_P_T, PLUS, D_E_P_T, PLUS, D_E_P_T},
		{D_E_P_T, PLUS, D_E_P_T, PLUS, D_T},
		{D_E_P_T, PLUS, D_T, PLUS, D_E_P_T},
		{D_E_P_T, PLUS, D_T, PLUS, D_T},
		{D_T, PLUS, D_E_P_T, PLUS, D_E_P_T},
		{D_T, PLUS, D_E_P_T, PLUS, D_T},
		{D_T, PLUS, D_T, PLUS, D_E_P_T},
		{D_T, PLUS, D_T, PLUS, D_T},
	}
	precMatrix := [][]gopapageno.Precedence{
		{gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecTakes, gopapageno.PrecTakes},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecTakes},
		{gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes},
	}
	bitPackedMatrix := []uint64{
		12130059261172884820, 160,
	}

	fn := func(rule uint16, lhs *gopapageno.Token, rhs []*gopapageno.Token, thread int) {
		var ruleType gopapageno.RuleFlags
		switch rule {
		case 0:
			ruleType = gopapageno.RuleSimple

			S0 := lhs
			D_E_P_T1 := rhs[0]

			S0.Child = D_E_P_T1
			S0.LastChild = D_E_P_T1

			{
				S0.Value = D_E_P_T1.Value
			}
			_ = D_E_P_T1
		case 1:
			ruleType = gopapageno.RuleSimple

			D_T0 := lhs
			D_E_P_T1 := rhs[0]
			DIVIDE2 := rhs[1]
			D_E_P_T3 := rhs[2]

			D_T0.Child = D_E_P_T1
			D_E_P_T1.Next = DIVIDE2
			DIVIDE2.Next = D_E_P_T3
			D_T0.LastChild = D_E_P_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_E_P_T1.Value.(*int64) / *D_E_P_T3.Value.(*int64)
				D_T0.Value = newValue
			}
			_ = D_E_P_T1
			_ = DIVIDE2
			_ = D_E_P_T3
		case 2:
			ruleType = gopapageno.RuleCyclic

			P0 := lhs
			D_E_P_T1 := rhs[0]
			PLUS2 := rhs[1]
			D_E_P_T3 := rhs[2]

			P0.Child = D_E_P_T1
			D_E_P_T1.Next = PLUS2
			PLUS2.Next = D_E_P_T3
			P0.LastChild = D_E_P_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_E_P_T1.Value.(*int64) + *D_E_P_T3.Value.(*int64)
				P0.Value = newValue
			}
			_ = D_E_P_T1
			_ = PLUS2
			_ = D_E_P_T3
		case 3:
			ruleType = gopapageno.RuleCyclic

			P0 := lhs
			D_E_P_T1 := rhs[0]
			PLUS2 := rhs[1]
			D_T3 := rhs[2]

			P0.Child = D_E_P_T1
			D_E_P_T1.Next = PLUS2
			PLUS2.Next = D_T3
			P0.LastChild = D_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_E_P_T1.Value.(*int64) + *D_T3.Value.(*int64)
				P0.Value = newValue
			}
			_ = D_E_P_T1
			_ = PLUS2
			_ = D_T3
		case 4:
			ruleType = gopapageno.RuleAppendLeft

			P0 := lhs
			D_E_P_T1 := rhs[0]
			PLUS2 := rhs[1]
			P3 := rhs[2]

			oldChild := P0
			P0.Child = D_E_P_T1
			D_E_P_T1.Next = PLUS2
			PLUS2.Next = P3
			P3.Next = oldChild

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_E_P_T1.Value.(*int64) + *P3.Value.(*int64)
				P0.Value = newValue
			}
			_ = D_E_P_T1
			_ = PLUS2
			_ = P3
		case 5:
			ruleType = gopapageno.RuleAppendLeft

			P0 := lhs
			D_E_P_T1 := rhs[0]
			PLUS2 := rhs[1]
			P3 := rhs[2]

			oldChild := P0
			P0.Child = D_E_P_T1
			D_E_P_T1.Next = PLUS2
			PLUS2.Next = P3
			P3.Next = oldChild

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_E_P_T1.Value.(*int64) + *P3.Value.(*int64)
				P0.Value = newValue
			}
			_ = D_E_P_T1
			_ = PLUS2
			_ = P3
		case 6:
			ruleType = gopapageno.RuleSimple

			S0 := lhs
			D_T1 := rhs[0]

			S0.Child = D_T1
			S0.LastChild = D_T1

			{
				S0.Value = D_T1.Value
			}
			_ = D_T1
		case 7:
			ruleType = gopapageno.RuleSimple

			D_T0 := lhs
			D_T1 := rhs[0]
			DIVIDE2 := rhs[1]
			D_E_P_T3 := rhs[2]

			D_T0.Child = D_T1
			D_T1.Next = DIVIDE2
			DIVIDE2.Next = D_E_P_T3
			D_T0.LastChild = D_E_P_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_T1.Value.(*int64) / *D_E_P_T3.Value.(*int64)
				D_T0.Value = newValue
			}
			_ = D_T1
			_ = DIVIDE2
			_ = D_E_P_T3
		case 8:
			ruleType = gopapageno.RuleCyclic

			P0 := lhs
			D_T1 := rhs[0]
			PLUS2 := rhs[1]
			D_E_P_T3 := rhs[2]

			P0.Child = D_T1
			D_T1.Next = PLUS2
			PLUS2.Next = D_E_P_T3
			P0.LastChild = D_E_P_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_T1.Value.(*int64) + *D_E_P_T3.Value.(*int64)
				P0.Value = newValue
			}
			_ = D_T1
			_ = PLUS2
			_ = D_E_P_T3
		case 9:
			ruleType = gopapageno.RuleCyclic

			P0 := lhs
			D_T1 := rhs[0]
			PLUS2 := rhs[1]
			D_T3 := rhs[2]

			P0.Child = D_T1
			D_T1.Next = PLUS2
			PLUS2.Next = D_T3
			P0.LastChild = D_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_T1.Value.(*int64) + *D_T3.Value.(*int64)
				P0.Value = newValue
			}
			_ = D_T1
			_ = PLUS2
			_ = D_T3
		case 10:
			ruleType = gopapageno.RuleAppendLeft

			P0 := lhs
			D_T1 := rhs[0]
			PLUS2 := rhs[1]
			P3 := rhs[2]

			oldChild := P0
			P0.Child = D_T1
			D_T1.Next = PLUS2
			PLUS2.Next = P3
			P3.Next = oldChild

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_T1.Value.(*int64) + *P3.Value.(*int64)
				P0.Value = newValue
			}
			_ = D_T1
			_ = PLUS2
			_ = P3
		case 11:
			ruleType = gopapageno.RuleAppendLeft

			P0 := lhs
			D_T1 := rhs[0]
			PLUS2 := rhs[1]
			P3 := rhs[2]

			oldChild := P0
			P0.Child = D_T1
			D_T1.Next = PLUS2
			PLUS2.Next = P3
			P3.Next = oldChild

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *D_T1.Value.(*int64) + *P3.Value.(*int64)
				P0.Value = newValue
			}
			_ = D_T1
			_ = PLUS2
			_ = P3
		case 12:
			ruleType = gopapageno.RuleSimple

			S0 := lhs
			P1 := rhs[0]

			S0.Child = P1
			S0.LastChild = P1

			{
				S0.Value = P1.Value
			}
			_ = P1
		case 13:
			ruleType = gopapageno.RuleAppend

			P0 := lhs
			P1 := rhs[0]
			PLUS2 := rhs[1]
			D_E_P_T3 := rhs[2]

			P0.LastChild.Next = PLUS2
			PLUS2.Next = D_E_P_T3
			P0.LastChild = D_E_P_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *P1.Value.(*int64) + *D_E_P_T3.Value.(*int64)
				P0.Value = newValue
			}
			_ = P1
			_ = PLUS2
			_ = D_E_P_T3
		case 14:
			ruleType = gopapageno.RuleAppend

			P0 := lhs
			P1 := rhs[0]
			PLUS2 := rhs[1]
			D_E_P_T3 := rhs[2]

			P0.LastChild.Next = PLUS2
			PLUS2.Next = D_E_P_T3
			P0.LastChild = D_E_P_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *P1.Value.(*int64) + *D_E_P_T3.Value.(*int64)
				P0.Value = newValue
			}
			_ = P1
			_ = PLUS2
			_ = D_E_P_T3
		case 15:
			ruleType = gopapageno.RuleAppend

			P0 := lhs
			P1 := rhs[0]
			PLUS2 := rhs[1]
			D_T3 := rhs[2]

			P0.LastChild.Next = PLUS2
			PLUS2.Next = D_T3
			P0.LastChild = D_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *P1.Value.(*int64) + *D_T3.Value.(*int64)
				P0.Value = newValue
			}
			_ = P1
			_ = PLUS2
			_ = D_T3
		case 16:
			ruleType = gopapageno.RuleAppend

			P0 := lhs
			P1 := rhs[0]
			PLUS2 := rhs[1]
			D_T3 := rhs[2]

			P0.LastChild.Next = PLUS2
			PLUS2.Next = D_T3
			P0.LastChild = D_T3

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *P1.Value.(*int64) + *D_T3.Value.(*int64)
				P0.Value = newValue
			}
			_ = P1
			_ = PLUS2
			_ = D_T3
		case 17:
			ruleType = gopapageno.RuleCombine

			P0 := lhs
			P1 := rhs[0]
			PLUS2 := rhs[1]
			P3 := rhs[2]

			P0.LastChild.Next = PLUS2
			PLUS2.Next = P3.Child
			P0.LastChild = P3.LastChild

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *P1.Value.(*int64) + *P3.Value.(*int64)
				P0.Value = newValue
			}
			_ = P1
			_ = PLUS2
			_ = P3
		case 18:
			ruleType = gopapageno.RuleCombine

			P0 := lhs
			P1 := rhs[0]
			PLUS2 := rhs[1]
			P3 := rhs[2]

			P0.LastChild.Next = PLUS2
			PLUS2.Next = P3.Child
			P0.LastChild = P3.LastChild

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *P1.Value.(*int64) + *P3.Value.(*int64)
				P0.Value = newValue
			}
			_ = P1
			_ = PLUS2
			_ = P3
		case 19:
			ruleType = gopapageno.RuleCombine

			P0 := lhs
			P1 := rhs[0]
			PLUS2 := rhs[1]
			P3 := rhs[2]

			P0.LastChild.Next = PLUS2
			PLUS2.Next = P3.Child
			P0.LastChild = P3.LastChild

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *P1.Value.(*int64) + *P3.Value.(*int64)
				P0.Value = newValue
			}
			_ = P1
			_ = PLUS2
			_ = P3
		case 20:
			ruleType = gopapageno.RuleCombine

			P0 := lhs
			P1 := rhs[0]
			PLUS2 := rhs[1]
			P3 := rhs[2]

			P0.LastChild.Next = PLUS2
			PLUS2.Next = P3.Child
			P0.LastChild = P3.LastChild

			{
				newValue := parserInt64Pools[thread].Get()
				*newValue = *P1.Value.(*int64) + *P3.Value.(*int64)
				P0.Value = newValue
			}
			_ = P1
			_ = PLUS2
			_ = P3
		case 21:
			ruleType = gopapageno.RuleSimple

			S0 := lhs
			S1 := rhs[0]

			S0.Child = S1
			S0.LastChild = S1

			{
				S0.Value = S1.Value
			}
			_ = S1
		case 22:
			ruleType = gopapageno.RuleSimple

			D_E_P_T0 := lhs
			LPAR1 := rhs[0]
			S2 := rhs[1]
			RPAR3 := rhs[2]

			D_E_P_T0.Child = LPAR1
			LPAR1.Next = S2
			S2.Next = RPAR3
			D_E_P_T0.LastChild = RPAR3

			{
				D_E_P_T0.Value = S2.Value
			}
			_ = LPAR1
			_ = S2
			_ = RPAR3
		case 23:
			ruleType = gopapageno.RuleSimple

			D_E_P_T0 := lhs
			NUMBER1 := rhs[0]

			D_E_P_T0.Child = NUMBER1
			D_E_P_T0.LastChild = NUMBER1

			{
				D_E_P_T0.Value = NUMBER1.Value
			}
			_ = NUMBER1
		}
		_ = ruleType
	}

	return &gopapageno.Grammar{
		NumTerminals:              numTerminals,
		NumNonterminals:           numNonTerminals,
		MaxRHSLength:              maxRHSLen,
		Rules:                     rules,
		CompressedRules:           compressedRules,
		PrecedenceMatrix:          precMatrix,
		BitPackedPrecedenceMatrix: bitPackedMatrix,
		MaxPrefixLength:           maxPrefixLength,
		Prefixes:                  prefixes,
		Func:                      fn,
		ParsingStrategy:           gopapageno.COPP,
		PreambleFunc:              ParserPreallocMem,
	}
}
