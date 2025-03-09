package xpath

import (
	"errors"
)

type contextSolutionsMap interface {
	addContextSolution(ctx NonTerminal, sols ...NonTerminal)
	transitiveClosure(maps ...contextSolutionsMap) contextSolutionsMap
	hasSolutionsFor(ctx NonTerminal) bool
	solutionsFor(ctx NonTerminal, maps ...contextSolutionsMap) []NonTerminal
	merge(incoming contextSolutionsMap) (result contextSolutionsMap, ok bool)
	convertToGroupOfSolutionsPositions() []Position
}

type contextSolutionEntry struct {
	solution NonTerminal
	next     *contextSolutionEntry
}

type contextSolutionsMapImpl struct {
	m map[NonTerminal]*contextSolutionEntry
}

func newContextSolutionsMap() contextSolutionsMap {
	return &contextSolutionsMapImpl{
		m: make(map[NonTerminal]*contextSolutionEntry),
	}
}

func (ctxSolMap *contextSolutionsMapImpl) addContextSolution(ctx NonTerminal, sols ...NonTerminal) {
	for _, sol := range sols {
		entry := contextSolutionEntry{solution: sol}
		entry.next = ctxSolMap.m[ctx]
		ctxSolMap.m[ctx] = &entry
	}
}

func (ctxSolMap *contextSolutionsMapImpl) hasSolutionsFor(ctx NonTerminal) bool {
	return len(ctxSolMap.solutionsFor(ctx)) > 0
}

func (ctxSolMap *contextSolutionsMapImpl) solutionsFor(ctx NonTerminal, maps ...contextSolutionsMap) (solutions []NonTerminal) {
	cur := ctxSolMap.m[ctx]
	for cur != nil {
		solutions = append(solutions, cur.solution)
		cur = cur.next
	}

	for currentMapIdx := 0; currentMapIdx < len(maps); currentMapIdx++ {
		tmpNodesToVisit := []NonTerminal{}

		for len(solutions) > 0 {
			currentNode := solutions[0]
			solutions = solutions[1:]
			tmpNodesToVisit = append(tmpNodesToVisit, maps[currentMapIdx].solutionsFor(currentNode)...)
		}
		solutions = tmpNodesToVisit
	}
	return
}

func (ctxSolMap *contextSolutionsMapImpl) transitiveClosure(maps ...contextSolutionsMap) (result contextSolutionsMap) {
	result = newContextSolutionsMap()

	for context := range ctxSolMap.m {
		solutionsReachableFromContext := ctxSolMap.solutionsFor(context, maps...)
		result.addContextSolution(context, solutionsReachableFromContext...)
	}
	return
}

func (ctxSolMap *contextSolutionsMapImpl) convertToGroupOfSolutionsPositions() (positions []Position) {
	for _, entry := range ctxSolMap.m {
		for entry != nil {
			positions = append(positions, entry.solution.Position())
			entry = entry.next
		}
	}
	return
}

func transitiveClosure(maps []contextSolutionsMap) contextSolutionsMap {
	start := maps[0]
	return start.transitiveClosure(maps[1:]...)
}

func (ctxSolMap *contextSolutionsMapImpl) merge(incoming contextSolutionsMap) (result contextSolutionsMap, ok bool) {
	result = ctxSolMap
	if incoming == nil {
		return
	}

	incomingImpl, ok := incoming.(*contextSolutionsMapImpl)

	if !ok {
		return
	}

	for k, v := range incomingImpl.m {
		cur := ctxSolMap.m[k]
		if cur == nil {
			ctxSolMap.m[k] = v
		} else {
			for cur.next != nil {
				cur = cur.next
			}
			cur.next = v
		}
		delete(incomingImpl.m, k)
	}
	ok = true
	return
}

// solutionsFor returns all the solutions that are reachable from the specified context
// by traversing all the contextSolutionsMaps which are passed as parameters
func solutionsFor(context NonTerminal, maps ...contextSolutionsMap) (solutions []NonTerminal, err error) {
	if context == nil {
		err = errors.New("context can NOT be nil")
		return
	}
	solutions = maps[0].solutionsFor(context)

	for currentMapIdx := 1; currentMapIdx < len(maps); currentMapIdx++ {
		tmpNodesToVisit := []NonTerminal{}

		for len(solutions) > 0 {
			currentNode := solutions[0]
			solutions = solutions[1:]
			tmpNodesToVisit = append(tmpNodesToVisit, maps[currentMapIdx].solutionsFor(currentNode)...)
		}
		solutions = tmpNodesToVisit
	}
	return
}
