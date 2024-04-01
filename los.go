package gopapageno

import (
	"fmt"
	"math"
)

// This is approx. 1MB per stack (on 64 bit architecture)
// TODO: This is like this for the original Symbol type.
const stackSize int = 26200

// stack contains a fixed size array of items,
// the current position in the stack
// and pointers to the previous and next stacks.
type stack[T any] struct {
	Data [stackSize]T
	Tos  int

	Prev *stack[T]
	Next *stack[T]
}

// LOS is a m of stacks.
// Each stack contains a fixed-size array of items.
type LOS[T any] struct {
	head *stack[T]
	cur  *stack[T]

	len  int
	pool *Pool[stack[T]]
}

// LOSIterator allows to iterate over a LOS, either forward or backwards.
type LOSIterator[T any] struct {
	los *LOS[T]

	cur *stack[T]
	pos int
}

// NewLOS creates a new LOS initialized with an empty stack.
func NewLOS[T any](pool *Pool[stack[T]]) *LOS[T] {
	s := pool.Get()

	return &LOS[T]{
		head: s,
		cur:  s,
		len:  0,
		pool: pool,
	}
}

func (l *LOS[T]) Push(t T) *T {
	// If the current stack is full, we must obtain a new one
	// and set it as the current one.
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

func (l *LOS[T]) Pop() *T {
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

// Merge links the stacks of the current and of another LOS.
func (l *LOS[T]) Merge(other *LOS[T]) {
	l.cur.Next = other.head
	other.head.Prev = l.cur

	l.cur = other.cur
	l.len += other.len
}

// Split splits a LOS into a slice of LOS of length n.
// The original LOS should not be used after this operation.
func (l *LOS[T]) Split(n int) ([]*LOS[T], error) {
	numStacks := l.NumStacks()

	if n > numStacks {
		return nil, fmt.Errorf("not enough stacks in LOS")
	}

	lists := make([]*LOS[T], n)
	curList := 0

	deltaStacks := float64(numStacks) / float64(n)
	assignedStacks := 0
	remainder := float64(0)

	curStack := l.head

	for assignedStacks < numStacks {
		remainder += deltaStacks

		stacksToAssign := int(math.Floor(remainder + 0.5))

		curStack.Prev = nil
		lists[curList] = &LOS[T]{
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

// NumStacks returns the number of stacks contained in the LOS.
// It takes linear time (in the number of stacks) to execute.
func (l *LOS[T]) NumStacks() int {
	n := 0

	for cur := l.head; cur != nil; cur = cur.Next {
		n++
	}
	return n
}

// Length returns the number of items contained in the LOS.
// It takes constant time to execute.
func (l *LOS[T]) Length() int {
	return l.len
}

// HeadIterator returns an iterator initialized to point before the first element of the m.
func (l *LOS[T]) HeadIterator() *LOSIterator[T] {
	return &LOSIterator[T]{l, l.head, -1}
}

// TailIterator returns an iterator initialized to point after the last element of the m.
func (l *LOS[T]) TailIterator() *LOSIterator[T] {
	return &LOSIterator[T]{l, l.cur, l.cur.Tos}
}

// Prev moves the iterator one position backward and returns a pointer to the new current element.
// It returns nil if it points before the first element of the m.
func (i *LOSIterator[T]) Prev() *T {
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
// It returns nil if it points before the first element or after the last element of the m.
func (i *LOSIterator[T]) Cur() *T {
	curStack := i.cur

	if i.pos >= 0 && i.pos < curStack.Tos {
		return &curStack.Data[i.pos]
	}

	return nil
}

// Next moves the iterator one position forward and returns a pointer to the new current element.
// It returns nil if it points after the last element of the m.
func (i *LOSIterator[T]) Next() *T {
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
