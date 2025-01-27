package xpath

import (
	"fmt"
)

// A new executionThread can be executed in the middle of a iteration
// over an execution thread list and does NOT cause the new execution
// thread to be considered by the running iteration.
type executionThread struct {
	offspr []executionThread
	ctx    *NonTerminal
	sol    *NonTerminal
	pp     pathPattern
	speculations swapbackArray[speculation]

	index int // index of the thread in executionThreadList
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

func (et *executionThread) checkAndUpdateSpeculations(v evaluator) (areSpeculationsFounded bool) {
	areSpeculationsFounded = true

	for _, speculation := range et.speculations.array[:et.speculations.size] {
		speculationValue := speculation.evaluate(v)
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

// concrete execution thread list implementation
type executionThreadList struct {
	list []executionThread
	size int
}
