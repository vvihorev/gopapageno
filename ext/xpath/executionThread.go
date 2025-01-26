package xpath

import (
	"container/list"
	"fmt"
)

// concrete execution thread implementation
type executionThread struct {
	offspr []executionThread
	ctx    NonTerminal
	sol    NonTerminal
	pp     pathPattern
	spList speculationList
	index int // index of the thread in executionThreadList
}

func (et *executionThread) String() string {
	return fmt.Sprintf("[ %v | %v | %v | %d ]", et.ctx, et.sol, et.pp, et.spList.len())
}

func (et *executionThread) setNTAsContextOrSolutionIfNotAlreadySet(contextOrSolution NonTerminal) {
	if et.ctx == nil {
		et.ctx = contextOrSolution
		return
	}

	if et.sol == nil {
		et.sol = contextOrSolution
	}
}

func (et *executionThread) context() NonTerminal {
	return et.ctx
}

func (et *executionThread) solution() NonTerminal {
	return et.sol
}

func (et *executionThread) pathPattern() pathPattern {
	return et.pp
}

func (et *executionThread) isCompleted() bool {
	return et.pp.isEmpty()
}

func (et *executionThread) isSpeculative() bool {
	return et.spList.len() != 0
}

func (et *executionThread) addSpeculation(prd predicate, ctx NonTerminal) speculation {
	return et.spList.addSpeculation(prd, ctx)
}

func (et *executionThread) removeSpeculation(sp speculation) {
	et.spList.removeSpeculation(sp)
}

func (et *executionThread) addChild(child executionThread) {
	et.offspr = append(et.offspr, child)
}

func (et *executionThread) children() []executionThread {
	return et.offspr
}

func (et *executionThread) checkAndUpdateSpeculations(v evaluator) (areSpeculationsFounded bool) {
	areSpeculationsFounded = true

	var next *list.Element
	for e := et.spList.actualList().Front(); e != nil; e = next {
		next = e.Next()
		speculation, ok := e.Value.(speculation)

		if !ok {
			panic(`speculation list iterate: can NOT access to the next speculation`)
		}

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

func (etList *executionThreadList) actualList() []executionThread {
	return etList.list
}

func newExecutionThreadList() executionThreadList {
	return executionThreadList{
		list: make([]executionThread, 0),
	}
}

// addExecutionThread adds a new execution thread to to the execution thread list.
// It can be executed in the middle of a iteration over an execution thread list and
// does NOT cause the new execution thread to be considered by the running iteration.
func (etList *executionThreadList) addExecutionThread(ctx, sol NonTerminal, pp pathPattern) executionThread {
	et := executionThread{
		ctx:    ctx,
		sol:    sol,
		pp:     pp,
		spList: newSpeculationList(),
	}
	etList.append(et)
	return et
}

func (etList *executionThreadList) append(et executionThread) {
	if etList.size > len(etList.list) {
		etList.list = append(etList.list, et)
	}
	etList.list[etList.size] = et
	et.index = etList.size
	etList.size++
}

func (etList *executionThreadList) removeExecutionThread(et executionThread, removeChildren bool) (ok bool) {
	if etList.size - 1 == et.index {
		etList.size--
		return
	}

	etList.list[et.index] = etList.list[etList.size]
	etList.list[et.index].index = et.index
	etList.size--

	if removeChildren {
		for _, childEt := range et.offspr {
			etList.removeExecutionThread(childEt, true)
		}
	}
	return
}

func (etList *executionThreadList) hasExecutionThreadRunningFor(ctx NonTerminal) bool {
	for i := 0; i < etList.size; i++ {
		found := etList.list[i].context() == ctx
		if found {
			return true
		}
	}
	return false
}

func (etList *executionThreadList) len() int {
	return etList.size
}

func (etList *executionThreadList) merge(incoming *executionThreadList) {
	if incoming != nil {
		for i := 0; i < incoming.size; i++ {
			etList.append(incoming.list[i])
		}
	}
}
