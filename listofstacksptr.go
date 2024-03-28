package gopapageno

import (
	"fmt"
	"math"
)

// This is approx. 1MB per stack (on 64 bit architecture)
const stackPtrSize int = 131000

type stackPtr struct {
	Data [stackSize]*Token
	Tos  int

	Prev *stackPtr
	Next *stackPtr
}

/*
listOfStackPtrs is a m of stacks containing Symbol pointers.
When the current stack is full a new one is automatically obtained and linked to it.
*/
type listOfStackPtrs struct {
	head          *stackPtr
	cur           *stackPtr
	len           int
	firstTerminal *Token
	pool          *Pool[stackPtr]
}

/*
iteratorPtr allows to iterate over a listOfStackPtrs, either forward or backward.
*/
type iteratorPtr struct {
	los *listOfStackPtrs
	cur *stackPtr
	pos int
}

/*
newLosPtr creates a new listOfStackPtrs initialized with one empty stack.
*/
func newLosPtr(pool *Pool[stackPtr]) listOfStackPtrs {
	firstStack := pool.Get()
	return listOfStackPtrs{firstStack, firstStack, 0, nil, pool}
}

/*
Push pushes a Symbol pointer in the listOfStackPtrs.
It returns the pointer itself.
*/
func (l *listOfStackPtrs) Push(sym *Token) *Token {
	curStack := l.cur

	//If the current stack is full obtain a new stack and set it as the current one
	if curStack.Tos < stackPtrSize {
		//Copy the Symbol pointer in the current position
		curStack.Data[curStack.Tos] = sym

		//If the Symbol is a terminal update the firstTerminal pointer
		if sym.Type.IsTerminal() {
			l.firstTerminal = sym
		}

		//Increment the current position
		curStack.Tos++

		//Increment the total length of the m
		l.len++

		//Return the Symbol pointer
		return sym
	}

	if curStack.Next != nil {
		curStack = curStack.Next
	} else {
		newStack := l.pool.Get()
		curStack.Next = newStack
		newStack.Prev = curStack
		curStack = newStack
	}
	l.cur = curStack

	//Copy the Symbol pointer in the current position
	curStack.Data[curStack.Tos] = sym

	//If the Symbol is a terminal update the firstTerminal pointer
	if sym.Type.IsTerminal() {
		l.firstTerminal = sym
	}

	//Increment the current position
	curStack.Tos++

	//Increment the total length of the m
	l.len++

	//Return the Symbol pointer
	return sym
}

/*
Pop pops a Symbol pointer from the stack and returns it.
*/
func (l *listOfStackPtrs) Pop() *Token {
	curStack := l.cur

	//Decrement the current position
	curStack.Tos--

	//While the current stack is empty set the previous stack as the current one.
	//If there are no more stacks return nil
	if curStack.Tos >= 0 {
		//Decrement the total length of the m
		l.len--

		//Return the Symbol pointer
		return curStack.Data[curStack.Tos]
	}

	curStack.Tos = 0

	if curStack.Prev == nil {
		return nil
	}

	curStack = curStack.Prev
	curStack.Tos--
	l.cur = curStack

	//Decrement the total length of the m
	l.len--

	//Return the Symbol pointer
	return curStack.Data[curStack.Tos]
}

/*
Merge merges a listOfStackPtrs to another by linking their stacks.
*/
func (l *listOfStackPtrs) Merge(l2 listOfStackPtrs) {
	l.cur.Next = l2.head
	l2.head.Prev = l.cur
	l.cur = l2.cur
	l.len += l2.len
	l.firstTerminal = l2.firstTerminal
}

/*
Split splits a listOfStackPtrs into a number of lists equal to numSplits,
which are returned as a slice of listOfStackPtrs.
If there are not at least numSplits stacks in the listOfStackPtrs it panics.
The original listOfStackPtrs should not be used after this operation.
*/
func (l *listOfStackPtrs) Split(numSplits int) []listOfStackPtrs {
	if numSplits > l.NumStacks() {
		panic(fmt.Sprintln("Cannot apply", numSplits, "splits on a ListOfStacks containing only", l.NumStacks(), "stacks."))
	}

	listsOfStacks := make([]listOfStackPtrs, numSplits)
	curList := 0

	numStacks := l.NumStacks()
	deltaStacks := float64(numStacks) / float64(numSplits)
	totAssignedStacks := 0
	remainder := float64(0)

	curStack := l.head

	for totAssignedStacks < numStacks {
		remainder += deltaStacks
		stacksToAssign := int(math.Floor(remainder + 0.5))

		curStack.Prev = nil
		listsOfStacks[curList] = listOfStackPtrs{curStack, curStack, curStack.Tos, nil, l.pool}

		for i := 1; i < stacksToAssign; i++ {
			curStack = curStack.Next
			listsOfStacks[curList].cur = curStack
			listsOfStacks[curList].len += curStack.Tos
		}
		nextStack := curStack.Next
		curStack.Next = nil
		curStack = nextStack

		listsOfStacks[curList].firstTerminal = listsOfStacks[curList].findFirstTerminal()

		remainder -= float64(stacksToAssign)
		totAssignedStacks += stacksToAssign

		curList++
	}

	return listsOfStacks
}

/*
Length returns the number of Symbol pointers contained in the listOfStackPtrs
*/
func (l *listOfStackPtrs) Length() int {
	return l.len
}

/*
NumStacks returns the number of stacks contained in the listOfStackPtrs
*/
func (l *listOfStackPtrs) NumStacks() int {
	i := 0

	curStack := l.head

	for curStack != nil {
		i++
		curStack = curStack.Next
	}

	return i
}

/*
FirstTerminal returns a pointer to the first terminal on the stack.
*/
func (l *listOfStackPtrs) FirstTerminal() *Token {
	return l.firstTerminal
}

/*
UpdateFirstTerminal should be used after a reduction in order to update the first terminal counter.
In fact, in order to save some time, only the Push operation automatically updates the first terminal pointer,
while the Pop operation does not.
*/
func (l *listOfStackPtrs) UpdateFirstTerminal() {
	l.firstTerminal = l.findFirstTerminal()
}

/*
This function is for internal usage only.
*/
func (l *listOfStackPtrs) findFirstTerminal() *Token {
	curStack := l.cur

	pos := curStack.Tos - 1

	for pos < 0 {
		pos = -1
		if curStack.Prev == nil {
			return nil
		}
		curStack = curStack.Prev
		pos = curStack.Tos - 1
	}

	for !curStack.Data[pos].Type.IsTerminal() {
		pos--
		for pos < 0 {
			pos = -1
			if curStack.Prev == nil {
				return nil
			}
			curStack = curStack.Prev
			pos = curStack.Tos - 1
		}
	}

	return curStack.Data[pos]
}

/*
Println prints the content of the listOfStackPtrs.
*/
func (l *listOfStackPtrs) Println() {
	iterator := l.HeadIterator()

	sym := iterator.Next()
	for sym != nil {
		fmt.Printf("(%d, %d) -> ", sym.Type, sym.Type)
		sym = iterator.Next()
	}
	fmt.Println()
}

/*
HeadIterator returns an Iterator starting at the first element of the m.
*/
func (l *listOfStackPtrs) HeadIterator() iteratorPtr {
	return iteratorPtr{l, l.head, -1}
}

/*
TailIterator returns an Iterator starting at the last element of the m.
*/
func (l *listOfStackPtrs) TailIterator() iteratorPtr {
	return iteratorPtr{l, l.cur, l.cur.Tos}
}

/*
Prev moves the ListIterator one position backward and returns a pointer to the current Symbol.
It returns nil if it points before the first element of the m.
*/
func (i *iteratorPtr) Prev() *Token {
	curStack := i.cur

	i.pos--

	if i.pos >= 0 {
		return curStack.Data[i.pos]
	}

	i.pos = -1
	if curStack.Prev == nil {
		return nil
	}
	curStack = curStack.Prev
	i.cur = curStack
	i.pos = curStack.Tos - 1

	return curStack.Data[i.pos]
}

/*
Cur returns a pointer to the current Symbol.
It returns nil if it points before the first element or after the last element of the m.
*/
func (i *iteratorPtr) Cur() *Token {
	curStack := i.cur

	if i.pos >= 0 && i.pos < curStack.Tos {
		return curStack.Data[i.pos]
	}

	return nil
}

/*
Next moves the ListIterator one position forward and returns a pointer to the current Symbol.
It returns nil if it points after the last element of the m.
*/
func (i *iteratorPtr) Next() *Token {
	curStack := i.cur

	i.pos++

	if i.pos < curStack.Tos {
		return curStack.Data[i.pos]
	}

	i.pos = curStack.Tos
	if curStack.Next == nil {
		return nil
	}
	curStack = curStack.Next
	i.cur = curStack
	i.pos = 0

	return curStack.Data[i.pos]
}
