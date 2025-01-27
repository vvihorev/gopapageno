package xpath

import (
	"errors"
)

type contextSolutionsMap map[*NonTerminal][]*NonTerminal

func (ctxSolMap contextSolutionsMap) addContextSolution(ctx *NonTerminal, sols ...*NonTerminal) {
	ctxSolMap[ctx] = append(ctxSolMap[ctx], sols...)
}

func (ctxSolMap contextSolutionsMap) hasSolutionsFor(ctx *NonTerminal) bool {
	return len(ctxSolMap.solutionsFor(ctx)) > 0
}

func (ctxSolMap contextSolutionsMap) solutionsFor(ctx *NonTerminal, maps ...contextSolutionsMap) (solutions []*NonTerminal) {
	solutions = ctxSolMap[ctx]

	for i := 0; i < len(maps); i++ {
		tmpNodesToVisit := []*NonTerminal{}

		for len(solutions) > 0 {
			currentNode := solutions[0]
			solutions = solutions[1:]
			tmpNodesToVisit = append(tmpNodesToVisit, maps[i].solutionsFor(currentNode)...)
		}
		solutions = tmpNodesToVisit
	}
	return
}

func (ctxSolMap contextSolutionsMap) transitiveClosure(maps ...contextSolutionsMap) (result contextSolutionsMap) {
	dest := make(contextSolutionsMap)

	for context := range ctxSolMap {
		solutionsReachableFromContext := ctxSolMap.solutionsFor(context, maps...)
		dest.addContextSolution(context, solutionsReachableFromContext...)
	}
	result = dest
	return
}

func (ctxSolMap contextSolutionsMap) convertToGroupOfSolutionsPositions() (positions []Position) {
	for _, solutions := range ctxSolMap {
		for _, solution := range solutions {
			positions = append(positions, solution.Position())
		}
	}
	return
}

func transitiveClosure(maps []contextSolutionsMap) contextSolutionsMap {
	start := maps[0]
	return start.transitiveClosure(maps[1:]...)
}

func (ctxSolMap contextSolutionsMap) merge(incoming contextSolutionsMap) {
	destination := ctxSolMap
	if incoming == nil {
		return
	}

	for k, v := range incoming {
		destination[k] = append(destination[k], v...)
	}
}

// solutionsFor returns all the solutions that are reachable from the specified context
// by traversing all the contextSolutionsMaps which are passed as parameters
func solutionsFor(context *NonTerminal, maps ...contextSolutionsMap) (solutions []*NonTerminal, err error) {
	if context == nil {
		err = errors.New("context can NOT be nil")
		return
	}
	solutions = maps[0].solutionsFor(context)

	for i := 1; i < len(maps); i++ {
		tmpNodesToVisit := []*NonTerminal{}

		for len(solutions) > 0 {
			currentNode := solutions[0]
			solutions = solutions[1:]
			tmpNodesToVisit = append(tmpNodesToVisit, maps[i].solutionsFor(currentNode)...)
		}
		solutions = tmpNodesToVisit
	}
	return
}
