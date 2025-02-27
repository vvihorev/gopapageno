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
	checkAndUpdateSpeculationsForReduction(r *Reduction) bool
}

// concrete execution thread implementation
type executionThreadImpl struct {
	ctx    NonTerminal
	sol    NonTerminal
	pp     pathPattern
	spList speculationList
	offspr []executionThread

	next *executionThreadImpl
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
	for e := et.spList.(*speculationListImpl).list.Front(); e != nil; e = next {
		next = e.Next()
		sp, ok := e.Value.(speculation)

		if !ok {
			panic(`speculation list iterate: can NOT access to the next speculation`)
		}

		speculationValue := sp.evaluate(v)
		switch speculationValue {
		case False:
			areSpeculationsFounded = false
			et.removeSpeculation(sp)
			break
		case True:
			et.removeSpeculation(sp)
		case Undefined:
		}
	}
	return
}

func (et *executionThreadImpl) checkAndUpdateSpeculationsForReduction(r *Reduction) (areSpeculationsFounded bool) {
	areSpeculationsFounded = true
	var next *list.Element
	for e := et.spList.(*speculationListImpl).list.Front(); e != nil; e = next {
		next = e.Next()
		sp, ok := e.Value.(speculation)

		if !ok {
			panic(`speculation list iterate: can NOT access to the next speculation`)
		}

		speculationValue := sp.evaluateReduction(r)
		switch speculationValue {
		case False:
			areSpeculationsFounded = false
			et.removeSpeculation(sp)
			break
		case True:
			et.removeSpeculation(sp)
		case Undefined:
		}
	}
	return
}

type executionThreadList interface {
	addExecutionThread(ctx, sol NonTerminal, pp pathPattern) executionThread
	removeExecutionThread(et executionThread, removeChildren bool) (ok bool)
	hasExecutionThreadRunningFor(ctx NonTerminal) bool
	iterate(callback executionThreadListIterableCallback)
	newIterator() executionThreadListIterator
	len() int
	merge(incoming executionThreadList) (result executionThreadList, ok bool)
}

// concrete execution thread list implementation
type executionThreadListImpl struct {
	head *executionThreadImpl
	size int
}

func newExecutionThreadList() executionThreadList {
	return &executionThreadListImpl{}
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
	et.next = etList.head
	etList.head = et
	etList.size++
	return et
}

func (etList *executionThreadListImpl) removeExecutionThread(et executionThread, removeChildren bool) (ok bool) {
	etImpl, ok := et.(*executionThreadImpl)
	if ok {
		prev := etList.head
		if prev == etImpl {
			etList.head = nil
		} else {
			cur := prev.next
			for cur != nil {
				if cur == etImpl {
					prev.next = cur.next
					break
				}
				prev = cur
				cur = cur.next
			}
		}
		etList.size--

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
		etImpl.next = nil   //avoid memory leaks
	}
	return
}

func (etList *executionThreadListImpl) hasExecutionThreadRunningFor(ctx NonTerminal) (found bool) {
	etList.iterate(func(et executionThread) (doBreak bool) {
		found = et.context() == ctx
		if found {
			doBreak = true
		}
		return
	})
	return
}

func (etList *executionThreadListImpl) len() int {
	return etList.size
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
	cur := etList.head
	if cur == nil {
		etList.head = incomingImpl.head
	} else {
		for ; cur.next != nil; cur = cur.next {
		}
		cur.next = incomingImpl.head
	}
	etList.size += incomingImpl.size

	// cleanup incoming execution thread list for reuse of executionTable
	incomingImpl.head = nil
	incomingImpl.size = 0
	return
}

// iterator object
func (etList *executionThreadListImpl) newIterator() executionThreadListIterator {
	return &executionThreadListIteratorImpl{
		nextEl: etList.head,
	}
}

type executionThreadListIterator interface {
	hasNext() bool
	next() (et executionThread, hasNext bool)
}

type executionThreadListIteratorImpl struct {
	nextEl *executionThreadImpl
}

func (etlIt *executionThreadListIteratorImpl) hasNext() bool {
	return etlIt.nextEl != nil
}

func (etlIt *executionThreadListIteratorImpl) next() (et executionThread, hasNext bool) {
	etlIt.nextEl = etlIt.nextEl.next
	hasNext = etlIt.nextEl != nil
	return
}

// iterate function
type executionThreadListIterableCallback func(et executionThread) (doBreak bool)

func (etList *executionThreadListImpl) iterate(callback executionThreadListIterableCallback) {
	for cur := etList.head; cur != nil; cur = cur.next {
		if doBreak := callback(cur); doBreak {
			return
		}
	}
}
