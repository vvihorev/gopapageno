package xpath

type storingIndex interface {
	setIndex(i int)
	getIndex() int
}

type swapbackArray[V storingIndex] struct {
	array []V
	size int
}

func (sa *swapbackArray[V]) append(v V) {
	if sa.size > len(sa.array) {
		sa.array = append(sa.array, v)
	}
	sa.array[sa.size] = v
	v.setIndex(sa.size)
	sa.size++
}

func (spl *swapbackArray[V]) remove(v V) {
	i := v.getIndex()
	if spl.size - 1 == i {
		spl.size--
		return
	}
	spl.size--
	spl.array[i] = spl.array[spl.size]
	spl.array[i].setIndex(i)
}
