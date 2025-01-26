package xpath

import (
	"fmt"
)

type speculation struct {
	evaluationsCount int
	prd              predicate
	ctx              NonTerminal
	index            int
}

func (s speculation) setIndex(i int) {
	s.index = i
}

func (s speculation) getIndex() int {
	return s.index
}

func (sp *speculation) String() string {
	return fmt.Sprintf("(%v , %v)", sp.ctx, sp.prd)
}

type evaluator func(id int, context NonTerminal, evaluationsCount int) customBool

func (sp *speculation) evaluate(v evaluator) (result customBool) {
	defer func() {
		sp.evaluationsCount++
	}()

	result = Undefined
	predicateAtomsIDs := sp.prd.atomsIDs()
	for _, atomID := range predicateAtomsIDs {
		id := int(atomID)
		atomValue := v(id, sp.ctx, sp.evaluationsCount)
		result = sp.prd.earlyEvaluate(atomID, atomValue)
		if result != Undefined {
			return result
		}
	}
	return
}
