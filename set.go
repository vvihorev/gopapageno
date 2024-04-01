package gopapageno

import (
	"cmp"
	"slices"
)

// Set is a collection that contains no duplicate elements.
type Set[T cmp.Ordered] struct {
	m map[T]struct{}
}

// NewSet returns a new empty set containing items of type T.
func NewSet[T cmp.Ordered]() *Set[T] {
	return &Set[T]{
		m: make(map[T]struct{}),
	}
}

func (s *Set[T]) Iter(yield func(int, T) bool) {
	i := 0
	for v := range s.m {
		if !yield(i, v) {
			return
		}
		i++
	}
}

// Contains returns whether v is contained within the set in constant time.
func (s *Set[T]) Contains(v T) bool {
	_, ok := s.m[v]
	return ok
}

// Add adds a new item to the set.
func (s *Set[T]) Add(v T) {
	s.m[v] = struct{}{}
}

// Remove removes a given item from the set.
// It results in a no-op if the element is not contained within the set.
func (s *Set[T]) Remove(v T) {
	delete(s.m, v)
}

// Clear removes all items from the set
func (s *Set[T]) Clear() {
	s.m = make(map[T]struct{})
}

// Len returns how many items are contained in the set.
func (s *Set[T]) Len() int {
	return len(s.m)
}

type FilterFunc[T comparable] func(v T) bool

// Filter returns a new set containing only the values that satisfy the predicate P.
func (s *Set[T]) Filter(P FilterFunc[T]) *Set[T] {
	res := NewSet[T]()
	for v := range s.m {
		if P(v) {
			res.Add(v)
		}
	}

	return res
}

func (s *Set[T]) Equals(o *Set[T]) bool {
	for k, v := range s.m {
		v2, ok := o.m[k]
		if !ok || v2 != v {
			return false
		}
	}

	for k2, v2 := range o.m {
		v, ok := s.m[k2]
		if !ok || v != v2 {
			return false
		}
	}

	return true
}

// Union returns a new Set containing all items from both s and o.
func (s *Set[T]) Union(o *Set[T]) *Set[T] {
	res := NewSet[T]()
	for v := range s.m {
		res.Add(v)
	}
	for v := range o.m {
		res.Add(v)
	}

	return res
}

// Intersection returns a new Set containing only common elements between s and o.
func (s *Set[T]) Intersection(o *Set[T]) *Set[T] {
	res := NewSet[T]()
	for v := range s.m {
		if o.Contains(v) {
			res.Add(v)
		}
	}

	return res
}

// Difference returns a new Set containing elements from s that are not contained in o.
func (s *Set[T]) Difference(o *Set[T]) *Set[T] {
	res := NewSet[T]()
	for v := range s.m {
		if !o.Contains(v) {
			res.Add(v)
		}
	}
	return res
}

// Copy returns a copy of s.
func (s *Set[T]) Copy() *Set[T] {
	res := NewSet[T]()
	for v := range s.m {
		res.Add(v)
	}

	return res
}

func (s *Set[T]) Slice() []T {
	slice := make([]T, len(s.m))
	i := 0
	for k, _ := range s.m {
		slice[i] = k
		i++
	}

	slices.Sort(slice)

	return slice
}
