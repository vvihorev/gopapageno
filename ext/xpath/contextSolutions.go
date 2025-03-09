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
	context, solution NonTerminal
	next              *contextSolutionEntry
}

type contextSolutionsMapImpl struct {
	head *contextSolutionEntry
}

func newContextSolutionsMap() contextSolutionsMap {
	return &contextSolutionsMapImpl{}
}

func (ctxSolMap *contextSolutionsMapImpl) addContextSolution(ctx NonTerminal, sols ...NonTerminal) {
	for _, sol := range sols {
		entry := contextSolutionEntry{context: ctx, solution: sol}
		entry.next = ctxSolMap.head
		ctxSolMap.head = &entry
	}
}

func (ctxSolMap *contextSolutionsMapImpl) hasSolutionsFor(ctx NonTerminal) bool {
	return len(ctxSolMap.solutionsFor(ctx)) > 0
}

func (ctxSolMap *contextSolutionsMapImpl) solutionsFor(ctx NonTerminal, maps ...contextSolutionsMap) (solutions []NonTerminal) {
	cur := ctxSolMap.head
	for cur != nil {
		if cur.context == ctx {
			solutions = append(solutions, cur.solution)
		}
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

	seen := map[NonTerminal]bool{}
	cur := ctxSolMap.head
	for cur != nil {
		if _, exists := seen[cur.context]; exists {
			continue
		}
		seen[cur.context] = true

		solutionsReachableFromContext := ctxSolMap.solutionsFor(cur.context, maps...)
		result.addContextSolution(cur.context, solutionsReachableFromContext...)
		cur = cur.next
	}
	return
}

func (ctxSolMap *contextSolutionsMapImpl) convertToGroupOfSolutionsPositions() (positions []Position) {
	cur := ctxSolMap.head
	for cur != nil {
		positions = append(positions, cur.solution.Position())
		cur = cur.next
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

	if incomingImpl.head == nil {
	} else if ctxSolMap.head == nil {
		ctxSolMap.head = incomingImpl.head
		incomingImpl.head = nil
	} else {
		cur := ctxSolMap.head
		for cur.next != nil {
			cur = cur.next
		}
		cur.next = incomingImpl.head
		incomingImpl.head = nil
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
