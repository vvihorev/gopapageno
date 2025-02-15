package xpath

import (
	"fmt"
	"strings"
)

type customBool int

// Custom boolean values which comprises the 'Undefined' value
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
	value       customBool
	root        *predNode
	undoneAtoms map[int]*predNode
}

type predNode struct {
	op     operator
	parent *predNode
	left   *predNode
	right  *predNode
}

func (p predicate) String() string {
	var rec func(n *predNode, sb *strings.Builder)
	rec = func(n *predNode, sb *strings.Builder) {
		sb.WriteString(fmt.Sprintf("%T [", n.op))
		if n.left != nil {
			rec(n.left, sb)
		}
		sb.WriteString("] [")
		if n.right != nil {
			rec(n.right, sb)
		}
		sb.WriteString("]")
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s: ", p.value.String()))
	rec(p.root, &sb)
	return sb.String()
}

/*
* This method returns the boolean value of the predicate as soon
* as it can be computed by means of the value assigned to a specific atom.
 */
func (p *predicate) earlyEvaluate(atomID int, value customBool) customBool {
	if p.value != Undefined || value == Undefined {
		return p.value
	}

	atom := p.undoneAtoms[atomID]
	delete(p.undoneAtoms, atomID)

	for value != Undefined && atom != nil {
		value = atom.op.evaluate(value)
		atom = atom.parent
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

func (n *predNode) copyNode(undoneSrc, undoneDst map[int]*predNode) *predNode {
	if n == nil {
		return nil
	}

	newNode := &predNode{
		op: n.op.copy(),
	}

	for k, v := range undoneSrc {
		if v == n {
			undoneDst[k] = newNode
		}
	}

	if n.left != nil {
		newNode.left = n.left.copyNode(undoneSrc, undoneDst)
		newNode.left.parent = newNode
	}
	if n.right != nil {
		newNode.right = n.right.copyNode(undoneSrc, undoneDst)
		newNode.right.parent = newNode
	}

	return newNode
}

func (p *predicate) copy() *predicate {
	if p == nil {
		return nil
	}

	undoneAtomsCopy := make(map[int]*predNode)
	newPredicate := &predicate{
		value: p.value,
		root:  p.root.copyNode(p.undoneAtoms, undoneAtomsCopy),
	}
	newPredicate.undoneAtoms = undoneAtomsCopy

	return newPredicate
}

type operator interface {
	evaluate(operand customBool) customBool
	copy() operator
}

type opConstructor func() operator

// atom operator concrete implementation
type atomOperator struct{}

func (op *atomOperator) copy() operator {
	return atom()
}

func (and *atomOperator) evaluate(operand customBool) customBool {
	return operand
}

func atom() operator {
	return new(atomOperator)
}

// not operator concrete implementation
type notOperator struct{}

func (op *notOperator) copy() operator {
	return not()
}

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

// or operator concrete implementation
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

func (op *orOperator) copy() operator {
	return or()
}

// and operator concrete implementation
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

func (op *andOperator) copy() operator {
	return and()
}

func and() operator {
	return new(andOperator)
}
