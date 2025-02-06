package xpath

import (
	"fmt"
)

// A new executionThread can be executed in the middle of a iteration
// over an execution thread list and does NOT cause the new execution
// thread to be considered by the running iteration.
type executionThread struct {
	offspr       []executionThread
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

func (et *executionThread) addSpeculation(prd *predicate, ctx NonTerminal) speculation {
	sp := speculation{prd: *prd, ctx: ctx}
	et.speculations.append(sp)
	return sp
}

func (et *executionThread) removeSpeculation(sp speculation) {
	et.speculations.remove(sp)
}

func (et *executionThread) addChild(child executionThread) {
	et.offspr = append(et.offspr, child)
}

func (et *executionThread) children() []executionThread {
	return et.offspr
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

	for _, speculation := range et.speculations.array[:et.speculations.size] {
		speculationValue := Undefined
		predicateAtomsIDs := speculation.prd.undoneAtoms
		for id, atom := range predicateAtomsIDs {
			// TODO(vvihorev): might be an imporer use of id here, it used to be the atomID that was passed
			atomValue := v(id, &speculation.ctx, speculation.evaluationsCount)
			speculationValue = speculation.prd.earlyEvaluate(atom, atomValue)
			if speculationValue != Undefined {
				break
			}
		}
		speculation.evaluationsCount++

		switch speculationValue {
		case False:
			areSpeculationsFounded = false
			et.removeSpeculation(speculation)
			continue
		case True:
			et.removeSpeculation(speculation)
		case Undefined:
		}
	}
	return
}
