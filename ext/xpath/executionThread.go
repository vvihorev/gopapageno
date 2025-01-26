package xpath

import (
	"container/list"
	"fmt"
)

type executionThread interface {
	context() NonTerminal
	solution() NonTerminal
	setNTAsContextOrSolutionIfNotAlreadySet(NonTerminal)
	pathPattern() pathPattern
	isCompleted() bool
	isSpeculative() bool
	addSpeculation(pr predicate, ctx NonTerminal) speculation
	removeSpeculation(sp speculation)
	addChild(et executionThread)
	children() []executionThread
	checkAndUpdateSpeculations(v evaluator) bool
}

// concrete execution thread implementation
type executionThreadImpl struct {
	ctx    NonTerminal
	sol    NonTerminal
	pp     pathPattern
	spList speculationList
	offspr []executionThread
	el     *list.Element
}

func (et *executionThreadImpl) String() string {
	return fmt.Sprintf("[ %v | %v | %v | %d ]", et.ctx, et.sol, et.pp, et.spList.len())
}

func (et *executionThreadImpl) setNTAsContextOrSolutionIfNotAlreadySet(contextOrSolution NonTerminal) {
	if et.ctx == nil {
		et.ctx = contextOrSolution
		return
	}

	if et.sol == nil {
		et.sol = contextOrSolution
	}
}

func (et *executionThreadImpl) context() NonTerminal {
	return et.ctx
}

func (et *executionThreadImpl) solution() NonTerminal {
	return et.sol
}

func (et *executionThreadImpl) pathPattern() pathPattern {
	return et.pp
}

func (et *executionThreadImpl) isCompleted() bool {
	return et.pp.isEmpty()
}

func (et *executionThreadImpl) isSpeculative() bool {
	return et.spList.len() != 0
}

func (et *executionThreadImpl) addSpeculation(prd predicate, ctx NonTerminal) speculation {
	return et.spList.addSpeculation(prd, ctx)
}

func (et *executionThreadImpl) removeSpeculation(sp speculation) {
	et.spList.removeSpeculation(sp)
}

func (et *executionThreadImpl) addChild(child executionThread) {
	et.offspr = append(et.offspr, child)
}

func (et *executionThreadImpl) children() []executionThread {
	return et.offspr
}

func (et *executionThreadImpl) checkAndUpdateSpeculations(v evaluator) (areSpeculationsFounded bool) {
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

type executionThreadList interface {
	addExecutionThread(ctx, sol NonTerminal, pp pathPattern) executionThread
	removeExecutionThread(et executionThread, removeChildren bool) (ok bool)
	hasExecutionThreadRunningFor(ctx NonTerminal) bool
	actualList() *list.List
	len() int
	merge(incoming executionThreadList) (result executionThreadList, ok bool)
}

// concrete execution thread list implementation
type executionThreadListImpl struct {
	list *list.List
}

func (etList *executionThreadListImpl) actualList() *list.List {
	return etList.list
}

func newExecutionThreadList() executionThreadList {
	return &executionThreadListImpl{
		list: list.New(),
	}
}

// addExecutionThread adds a new execution thread to to the execution thread list.
// It can be executed in the middle of a iteration over an execution thread list and
// does NOT cause the new execution thread to be considered by the running iteration.
func (etList *executionThreadListImpl) addExecutionThread(ctx, sol NonTerminal, pp pathPattern) executionThread {
	et := &executionThreadImpl{
		ctx:    ctx,
		sol:    sol,
		pp:     pp,
		spList: newSpeculationList(),
	}
	et.el = etList.list.PushFront(et)
	return et
}

func (etList *executionThreadListImpl) removeExecutionThread(et executionThread, removeChildren bool) (ok bool) {
	etImpl, ok := et.(*executionThreadImpl)
	if ok {
		etList.list.Remove(etImpl.el)
		if removeChildren {
			for _, childEt := range etImpl.offspr {
				etList.removeExecutionThread(childEt, true)
			}
		}
		etImpl.ctx = nil    //avoid memory leaks
		etImpl.sol = nil    //avoid memory leaks
		etImpl.pp = nil     //avoid memory leaks
		etImpl.spList = nil //avoid memory leaks
		etImpl.offspr = nil //avoid memory leaks
		etImpl.el = nil     //avoid memory leaks
	}
	return
}

func (etList *executionThreadListImpl) hasExecutionThreadRunningFor(ctx NonTerminal) (found bool) {
	var next *list.Element
	for e := etList.list.Front(); e != nil; e = next {
		next = e.Next()
		et, ok := e.Value.(executionThread)
		if !ok {
			panic(`execution thread list iterate: can NOT access to the next execution thread`)
		}

		found = et.context() == ctx
		if found {
			break
		}
	}
	return
}

func (etList *executionThreadListImpl) len() int {
	return etList.list.Len()
}

func (etList *executionThreadListImpl) merge(incoming executionThreadList) (result executionThreadList, ok bool) {
	result = etList
	if incoming == nil {
		return
	}
	incomingImpl, ok := incoming.(*executionThreadListImpl)
	if !ok {
		return
	}

	ok = true
	etList.list.PushBackList(incomingImpl.list)
	for el := etList.list.Front(); el != nil; el = el.Next() {
		incomingEt := el.Value.(*executionThreadImpl)
		incomingEt.el = el
	}
	return
}
