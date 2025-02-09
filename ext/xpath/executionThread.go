package xpath

import (
	"fmt"
)

// A new executionThread can be executed in the middle of a iteration
// over an execution thread list and does NOT cause the new execution
// thread to be considered by the running iteration.
type executionThread struct {
	ctx          *NonTerminal
	sol          *NonTerminal
	pp           pathPattern
	speculations swapbackArray[speculation]

	index int // index of the thread in executionThreadList
}

// concrete execution thread list implementation
type executionThreadList struct {
	list []executionThread
	size int
}

type speculation struct {
	evaluationsCount int
	prd              predicate
	ctx              NonTerminal
	index            int
}

func (et executionThread) getIndex() int {
	return et.index
}

func (et executionThread) setIndex(i int) {
	et.index = i
}

func (et *executionThread) String() string {
	return fmt.Sprintf("[ %v | %v | %v | %d ]", et.ctx, et.sol, et.pp, et.speculations.size)
}

func (et *executionThread) isSpeculative() bool {
	return et.speculations.size != 0
}

func (et *executionThread) addSpeculation(prd *predicate, ctx NonTerminal) {
	sp := speculation{prd: *prd, ctx: ctx}
	et.speculations.append(sp)
}

func (et *executionThread) removeSpeculation(sp speculation) {
	et.speculations.remove(sp)
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

type evaluator func(id int, context *NonTerminal, evaluationsCount int) customBool

func (et *executionThread) checkAndUpdateSpeculations(v evaluator) (areSpeculationsFounded bool) {
	areSpeculationsFounded = true

	i := 0
	for i < et.speculations.size {
		speculationValue := Undefined
		predicateAtomsIDs := et.speculations.array[i].prd.undoneAtoms
		for atomID := range predicateAtomsIDs {
			atomValue := v(atomID, &et.speculations.array[i].ctx, et.speculations.array[i].evaluationsCount)
			speculationValue = et.speculations.array[i].prd.earlyEvaluate(atomID, atomValue)
			if speculationValue != Undefined {
				break
			}
		}
		et.speculations.array[i].evaluationsCount++

		switch speculationValue {
		case False:
			areSpeculationsFounded = false
			et.removeSpeculation(et.speculations.array[i])
			// TODO(vvihorev): need to double check if this should break
			break
		case True:
			et.removeSpeculation(et.speculations.array[i])
		case Undefined:
			i++
		}
	}
	return
}
