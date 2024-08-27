package gopapageno

import (
	"math"
	"reflect"
)

// stack contains a fixed size array of items,
// the current position in the stack and pointers to the previous and next stacks.
type stack[T any] struct {
	// Data [stackSize]T
	Data []T
	Tos  int
	Size int

	Prev *stack[T]
	Next *stack[T]
}

func newStackFactory[T any](length int) func() *stack[T] {
	return func() *stack[T] {
		return &stack[T]{
			Data: make([]T, length),
			Size: length,
		}
	}
}

func stackLengthFor[T any](sizeFactor float64) int {
	return int(float64(stackSize[T]()) * sizeFactor)
}

func newStack[T any]() *stack[T] {
	stackLen := stackSize[T]()

	return &stack[T]{
		Data: make([]T, stackLen),
		Size: stackLen,
	}
}

func newStackOf[T any](length int) *stack[T] {
	return &stack[T]{
		Data: make([]T, length),
		Size: length,
	}
}

func (s *stack[T]) Push(t T) {
	if s.Tos >= s.Size {
		panic("calculations were wrong.")
	}

	s.Data[s.Tos] = t

	s.Tos++
}

func (s *stack[T]) Replace(t T) {
	if s.Tos == 0 {
		panic("calculations were wrong.")
	}

	s.Data[s.Tos-1] = t
}

func (s *stack[T]) Slice(from int, length int) []T {
	return s.Data[from : from+length]
}

func stackSize[T any]() int {
	typeSize := reflect.TypeFor[T]().Size()
	return 1024 * 1024 / int(typeSize)
}

func stacksCount[T any](src []byte, concurrency int, avgTokenLen int) int {
	return int(math.Ceil(float64(len(src)) / float64(avgTokenLen) / float64(concurrency) / float64(stackSize[T]())))
}

func stacksCountFactored[T any](src []byte, opts *RunOptions) int {
	parallelMult := 1.0 - (0.999 * opts.ParallelFactor)
	elements := float64(len(src)) / float64(opts.AvgTokenLength) / float64(opts.Concurrency) * parallelMult

	return int(math.Floor(elements / float64(stackSize[T]())))
}
