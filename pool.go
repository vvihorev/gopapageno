package gopapageno

// A Pool can be used to preallocate a number of items of type T.
// It is not thread-safe.
type Pool[T any] struct {
	pool []T
	cur  int
}

// NewPool creates a new pool, allocating `length` elements.
func NewPool[T any](length int) *Pool[T] {
	return &Pool[T]{make([]T, length), 0}
}

// Get returns an item from the pool if available. Otherwise, it initializes a new one.
// It is not thread-safe.
func (p *Pool[T]) Get() *T {
	if p.cur >= len(p.pool) {
		return new(T)
	}

	addr := &p.pool[p.cur]
	p.cur++

	return addr
}

// Left returns the number of items remaining in the pool.
func (p *Pool[T]) Left() int {
	return len(p.pool) - p.cur
}
