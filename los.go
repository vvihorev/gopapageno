package gopapageno

import (
	"fmt"
	"math"
)

// This is 1MB per stack (on 64 bit architecture)
const stackSize int = 26214

// stack contains a fixed size array of items,
// the current position in the stack
// and pointers to the previous and next stacks.
type stack[T any] struct {
	Data [stackSize]T
	Tos  int

	Prev *stack[T]
	Next *stack[T]
}

// listOfStacks is a list of stacks.
type listOfStacks[T any] struct {
	head *stack[T]
	cur  *stack[T]

	len  int
	pool *Pool[stack[T]]
}

// losIterator allows to iterate over a listOfStacks, either forward or backwards.
type losIterator[T any] struct {
	los *listOfStacks[T]

	cur *stack[T]
	pos int
}

// newListOfStacks creates a new listOfStacks initialized with an empty stack.
func newListOfStacks[T any](pool *Pool[stack[T]]) *listOfStacks[T] {
	s := pool.Get()

	return &listOfStacks[T]{
		head: s,
		cur:  s,
		len:  0,
		pool: pool,
	}
}

// Push adds an element to the listOfStacks.
// By default, the element is added to the current stack;
// if that is full, a new one is obtained from the pool.
func (l *listOfStacks[T]) Push(t T) *T {
	// If the current stack is full, we must obtain a new one and set it as the current one.
	if l.cur.Tos >= stackSize {
		if l.cur.Next != nil {
			l.cur = l.cur.Next
		} else {
			s := l.pool.Get()

			l.cur.Next = s
			s.Prev = l.cur

			l.cur = s
		}
	}

	l.cur.Data[l.cur.Tos] = t
	ptr := &l.cur.Data[l.cur.Tos]

	l.cur.Tos++
	l.len++

	return ptr
}

// Pop removes the topmost element from the listOfStacks and returns it.
func (l *listOfStacks[T]) Pop() *T {
	l.cur.Tos--

	if l.cur.Tos >= 0 {
		l.len--
		return &l.cur.Data[l.cur.Tos]
	}

	l.cur.Tos = 0

	if l.cur.Prev == nil {
		return nil
	}

	l.cur = l.cur.Prev
	l.cur.Tos--
	l.len--

	return &l.cur.Data[l.cur.Tos]
}

// Merge links the stacks of the current and of another listOfStacks.
func (l *listOfStacks[T]) Merge(other *listOfStacks[T]) {
	l.cur.Next = other.head
	other.head.Prev = l.cur

	l.cur = other.cur
	l.len += other.len
}

// Split splits a listOfStacks into a slice of listOfStacks of length n.
// The original listOfStacks should not be used after this operation.
func (l *listOfStacks[T]) Split(n int) ([]*listOfStacks[T], error) {
	numStacks := l.NumStacks()

	if n > numStacks {
		return nil, fmt.Errorf("not enough stacks in listOfStacks")
	}

	lists := make([]*listOfStacks[T], n)
	curList := 0

	deltaStacks := float64(numStacks) / float64(n)
	assignedStacks := 0
	remainder := float64(0)

	curStack := l.head

	for assignedStacks < numStacks {
		remainder += deltaStacks

		stacksToAssign := int(math.Floor(remainder + 0.5))

		curStack.Prev = nil
		lists[curList] = &listOfStacks[T]{
			head: curStack,
			cur:  curStack,
			len:  curStack.Tos,
			pool: l.pool,
		}

		for i := 1; i < stacksToAssign; i++ {
			curStack = curStack.Next
			lists[curList].cur = curStack
			lists[curList].len += curStack.Tos
		}

		next := curStack.Next
		curStack.Next = nil

		curStack = next

		remainder -= float64(stacksToAssign)
		assignedStacks += stacksToAssign

		curList++
	}

	return lists, nil
}

// NumStacks returns the number of stacks contained in the listOfStacks.
// It takes linear time (in the number of stacks) to execute.
func (l *listOfStacks[T]) NumStacks() int {
	n := 0

	for cur := l.head; cur != nil; cur = cur.Next {
		n++
	}
	return n
}

// Length returns the number of items contained in the listOfStacks.
// It takes constant time to execute.
func (l *listOfStacks[T]) Length() int {
	return l.len
}

// HeadIterator returns an iterator initialized to point before the first element of the list.
func (l *listOfStacks[T]) HeadIterator() *losIterator[T] {
	return &losIterator[T]{l, l.head, -1}
}

// TailIterator returns an iterator initialized to point after the last element of the list.
func (l *listOfStacks[T]) TailIterator() *losIterator[T] {
	return &losIterator[T]{l, l.cur, l.cur.Tos}
}

// Prev moves the iterator one position backward and returns a pointer to the new current element.
// It returns nil if it points before the first element of the list.
func (i *losIterator[T]) Prev() *T {
	curStack := i.cur

	i.pos--

	if i.pos >= 0 {
		return &curStack.Data[i.pos]
	}

	i.pos = -1
	if curStack.Prev == nil {
		return nil
	}
	curStack = curStack.Prev
	i.cur = curStack
	i.pos = curStack.Tos - 1

	return &curStack.Data[i.pos]
}

// Cur returns a pointer to the current element.
// It returns nil if it points before the first element or after the last element of the list.
func (i *losIterator[T]) Cur() *T {
	curStack := i.cur

	if i.pos >= 0 && i.pos < curStack.Tos {
		return &curStack.Data[i.pos]
	}

	return nil
}

// Next moves the iterator one position forward and returns a pointer to the new current element.
// It returns nil if it points after the last element of the list.
func (i *losIterator[T]) Next() *T {
	curStack := i.cur

	i.pos++

	if i.pos < curStack.Tos {
		return &curStack.Data[i.pos]
	}

	i.pos = curStack.Tos
	if curStack.Next == nil {
		return nil
	}
	curStack = curStack.Next
	i.cur = curStack
	i.pos = 0

	return &curStack.Data[i.pos]
}

// This is 1MB per stack (on 64 bit architecture)
const pointerStackSize int = 131072

// tokenPointerStack contains a fixed size array of pointer to Tokens,
// the current position in the stack
// and pointers to the previous and next stacks.
type tokenPointerStack struct {
	Data [pointerStackSize]*Token
	Tos  int

	Prev *tokenPointerStack
	Next *tokenPointerStack
}

// listOfStacks is a list of pointer stacks.
type listOfTokenPointerStacks struct {
	head          *tokenPointerStack
	cur           *tokenPointerStack
	len           int
	firstTerminal *Token
	pool          *Pool[tokenPointerStack]
}

/*
listOfTokenPointerStacksIterator allows to iterate over a listOfTokenPointerStacks, either forward or backward.
*/
type listOfTokenPointerStacksIterator struct {
	los *listOfTokenPointerStacks

	cur *tokenPointerStack
	pos int
}

// newListOfTokenPointerStacks creates a new listOfTokenPointerStacks initialized with one empty stack.
func newListOfTokenPointerStacks(pool *Pool[tokenPointerStack]) *listOfTokenPointerStacks {
	s := pool.Get()

	return &listOfTokenPointerStacks{
		s,
		s,
		0,
		nil,
		pool,
	}
}

// FirstTerminal returns a pointer to the first terminal on the stack.
func (l *listOfTokenPointerStacks) FirstTerminal() *Token {
	return l.firstTerminal
}

// UpdateFirstTerminal should be used after a reduction in order to update the first terminal counter.
// In fact, in order to save some time, only the Push operation automatically updates the first terminal pointer,
// while the Pop operation does not.
func (l *listOfTokenPointerStacks) UpdateFirstTerminal() {
	l.firstTerminal = l.findFirstTerminal()
}

// findFirstTerminals computes the first terminal on the stacks.
// This function is for internal usage only.
func (l *listOfTokenPointerStacks) findFirstTerminal() *Token {
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

// Push pushes a token pointer in the listOfTokenPointerStacks.
// It returns the pointer itself.
func (l *listOfTokenPointerStacks) Push(token *Token) *Token {
	// If the current stack is full, we must obtain a new one and set it as the current one.
	if l.cur.Tos >= stackSize {
		if l.cur.Next != nil {
			l.cur = l.cur.Next
		} else {
			s := l.pool.Get()

			l.cur.Next = s
			s.Prev = l.cur

			l.cur = s
		}
	}

	l.cur.Data[l.cur.Tos] = token

	//If the token is a terminal update the firstTerminal pointer
	if token.Type.IsTerminal() {
		l.firstTerminal = token
	}

	l.cur.Tos++
	l.len++

	return token
}

// Pop removes the topmost element from the listOfTokenPointerStacks and returns it.
func (l *listOfTokenPointerStacks) Pop() *Token {
	l.cur.Tos--

	if l.cur.Tos >= 0 {
		l.len--
		return l.cur.Data[l.cur.Tos]
	}

	l.cur.Tos = 0

	if l.cur.Prev == nil {
		return nil
	}

	l.cur = l.cur.Prev
	l.cur.Tos--
	l.len--

	return l.cur.Data[l.cur.Tos]
}

// Merge links the stacks of the current and of another listOfTokenPointerStacks.
func (l *listOfTokenPointerStacks) Merge(other *listOfTokenPointerStacks) {
	l.cur.Next = other.head
	other.head.Prev = l.cur

	l.cur = other.cur
	l.len += other.len

	l.firstTerminal = other.firstTerminal
}

// Split splits a listOfStacks into a slice of listOfStacks of length n.
// The original listOfStacks should not be used after this operation.
func (l *listOfTokenPointerStacks) Split(n int) ([]*listOfTokenPointerStacks, error) {
	numStacks := l.NumStacks()

	if n > numStacks {
		return nil, fmt.Errorf("not enough stacks in listOfTokenPointerStacks")
	}

	lists := make([]*listOfTokenPointerStacks, n)
	curList := 0

	deltaStacks := float64(numStacks) / float64(n)
	assignedStacks := 0
	remainder := float64(0)

	curStack := l.head

	for assignedStacks < numStacks {
		remainder += deltaStacks

		stacksToAssign := int(math.Floor(remainder + 0.5))

		curStack.Prev = nil
		lists[curList] = &listOfTokenPointerStacks{
			head: curStack,
			cur:  curStack,
			len:  curStack.Tos,
			pool: l.pool,
		}

		for i := 1; i < stacksToAssign; i++ {
			curStack = curStack.Next
			lists[curList].cur = curStack
			lists[curList].len += curStack.Tos
		}

		next := curStack.Next
		curStack.Next = nil

		curStack = next

		lists[curList].firstTerminal = lists[curList].findFirstTerminal()

		remainder -= float64(stacksToAssign)
		assignedStacks += stacksToAssign

		curList++
	}

	return lists, nil
}

// NumStacks returns the number of stacks contained in the listOfTokenPointerStacks.
// It takes linear time (in the number of stacks) to execute.
func (l *listOfTokenPointerStacks) NumStacks() int {
	n := 0

	for cur := l.head; cur != nil; cur = cur.Next {
		n++
	}
	return n
}

// Length returns the number of items contained in the listOfTokenPointerStacks.
// It takes constant time to execute.
func (l *listOfTokenPointerStacks) Length() int {
	return l.len
}

// HeadIterator returns an iterator initialized to point before the first element of the list.
func (l *listOfTokenPointerStacks) HeadIterator() *listOfTokenPointerStacksIterator {
	return &listOfTokenPointerStacksIterator{l, l.head, -1}
}

// TailIterator returns an iterator initialized to point after the last element of the list.
func (l *listOfTokenPointerStacks) TailIterator() *listOfTokenPointerStacksIterator {
	return &listOfTokenPointerStacksIterator{l, l.cur, l.cur.Tos}
}

// Prev moves the iterator one position backward and returns a pointer to the new current element.
// It returns nil if it points before the first element of the list.
func (i *listOfTokenPointerStacksIterator) Prev() *Token {
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

// Cur returns a pointer to the current element.
// It returns nil if it points before the first element or after the last element of the list.
func (i *listOfTokenPointerStacksIterator) Cur() *Token {
	curStack := i.cur

	if i.pos >= 0 && i.pos < curStack.Tos {
		return curStack.Data[i.pos]
	}

	return nil
}

// Next moves the iterator one position forward and returns a pointer to the new current element.
// It returns nil if it points after the last element of the list.
func (i *listOfTokenPointerStacksIterator) Next() *Token {
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
