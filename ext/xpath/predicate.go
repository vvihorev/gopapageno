package xpath

import (
	"fmt"
	"math"
)

type customBool int

//Custom boolean values which comprises the 'Undefined' value
const (
	Undefined customBool = iota
	False
	True
)

func (cb customBool) String() string {
	switch cb {
	case Undefined:
		return "Undefined"
	case True:
		return "True"
	case False:
		return "False"
	default:
		return "Undefined"
	}
}

type atomID int

type predicate struct {
	value            customBool
	expressionVector []operator
	atomsLookup      map[atomID]int
}

func newPredicate() *predicate {
	return new(predicate)
}

func (p *predicate) String() string {
	return fmt.Sprintf("p[%v]", p.atomsIDs())
}

/*
* parentIndexOf computes the index of the parent operator
* inside the flat representation of the predicate's binary
* tree data structure.
 */
func (p *predicate) parentIndexOf(opIndex int) int {
	if opIndex == 0 {
		return -1
	}
	return int(math.Floor(float64((opIndex - 1) / 2)))
}

/*
* leftChildIndexOf computes the index of the left child operator
* inside the flat representation of the predicate's binary
* tree data structure.
 */
func (p *predicate) leftChildIndexOf(opIndex int) int {
	return 2*opIndex + 1
}

/*
* rightChildIndexOf computes the index of the right child operator
* inside the flat representation of the predicate's binary
* tree data structure.
 */
func (p *predicate) rightChildIndexOf(opIndex int) int {
	return 2*opIndex + 2
}

/*
* atomsIDs returns all the atoms which take part to the predicate.
* The order by which the atomID appear does NOT respect the order by
* which they were added to the predicate
 */
func (p *predicate) atomsIDs() []atomID {
	keys := make([]atomID, 0, len(p.atomsLookup))
	for k := range p.atomsLookup {
		keys = append(keys, k)
	}
	return keys
}

/*
* This method returns the boolean value of the predicate as soon
* as it can be computed by means of the value assigned to a specific atom.
 */
func (p *predicate) earlyEvaluate(atomID atomID, value customBool) customBool {
	currentOpIndex, ok := p.atomsLookup[atomID]
	if p.value != Undefined || !ok || value == Undefined {
		return p.value
	}
	delete(p.atomsLookup, atomID)

	for value != Undefined && currentOpIndex >= 0 {
		currentOp := p.expressionVector[currentOpIndex]
		value = currentOp.evaluate(value)
		currentOpIndex = p.parentIndexOf(currentOpIndex)
	}
	p.value = value
	return value
}

func toCustomBool(b bool) customBool {
	switch b {
	case true:
		return True
	case false:
		return False
	default:
		panic("from boolean to customBool conversion error: received an impossible boolean")
	}
}

func (cb customBool) tobool() (value, ok bool) {
	switch cb {
	case True:
		value = true
		ok = true
	case False:
		value = false
		ok = true
	default:
		value = false
		ok = false
	}
	return
}

func (p *predicate) copy() *predicate {
	expressionVectorCopy := make([]operator, len(p.expressionVector))
	copy(expressionVectorCopy, p.expressionVector)

	atomsLookupCopy := make(map[atomID]int)
	for k, v := range p.atomsLookup {
		atomsLookupCopy[k] = v
	}

	return &predicate{
		expressionVector: expressionVectorCopy,
		atomsLookup:      atomsLookupCopy,
	}
}

type operator interface {
	evaluate(operand customBool) customBool
}

type opConstructor func() operator

//atom operator concrete implementation
type atomOperator struct{}

func (and *atomOperator) evaluate(operand customBool) customBool {
	return operand
}

func atom() operator {
	return new(atomOperator)
}

//not operator concrete implementation
type notOperator struct {}

func (not *notOperator) evaluate(operand customBool) customBool {
	switch operand {
	case True:
		return False
	case False:
		return True
	default:
		return Undefined
	}
}

func not() operator {
	return new(notOperator)
}

//or operator concrete implementation
type orOperator struct {
	previousOperand customBool
}

func (or *orOperator) evaluate(operand customBool) customBool {
	if operand == True {
		return True
	}
	if operand == False {
		if or.previousOperand == False {
			return False
		}
		or.previousOperand = False
	}
	return Undefined
}

func or() operator {
	return new(orOperator)
}

//and operator concrete implementation
type andOperator struct {
	previousOperand customBool
}

func (and *andOperator) evaluate(operand customBool) customBool {
	if operand == False {
		return False
	}
	if operand == True {
		if and.previousOperand == True {
			return True
		}
		and.previousOperand = True
	}
	return Undefined
}

func and() operator {
	return new(andOperator)
}
