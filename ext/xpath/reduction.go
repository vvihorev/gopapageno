package xpath

import (
	"fmt"
)

type Reduction struct {
	reducedNT, generativeNT, wrappedNT *NonTerminal
	updatingExecutionTable             executionTable
	globalUdpeRecordBeingConsidered    *globalUdpeRecord
}

func (r *Reduction) Setup(reducedNT, generativeNT, wrappedNT *NonTerminal) {
	var updatingExecutionTable executionTable

	if wrappedNT != nil {
		updatingExecutionTable = *wrappedNT.executionTable
	} else {
		updatingExecutionTable = udpeGlobalTable.newExecutionTable()
	}

	r.reducedNT = reducedNT
	r.generativeNT = generativeNT
	r.wrappedNT = wrappedNT
	r.updatingExecutionTable = updatingExecutionTable
}

func (r *Reduction) Handle() {
	logger.Printf("REDUCTION: %v <- %v <> %v </>", r.reducedNT, r.generativeNT, r.wrappedNT)
	r.iterateOverAllGlobalUdpeRecordsAndExecuteMainPhases()
	r.prepareUpdatingExecutionTableToBePropagatedToReducedNT()
	logger.Printf("%v", r.updatingExecutionTable.String())
	r.propagateUpdatingExecutionTableToReducedNT()
}

func (r *Reduction) iterateOverAllGlobalUdpeRecordsAndExecuteMainPhases() {
	for id, gr := range udpeGlobalTable.list {
		r.globalUdpeRecordBeingConsidered = &gr

		r.addNewExecutionThreadsToExecutionRecord(&r.updatingExecutionTable[id])

		r.updatingExecutionTable[id].updateAllExecutionThreads(r.reducedNT)
		r.updatingExecutionTable[id].stopUnfoundedSpeculativeExecutionThreads(r.updatingExecutionTable.evaluateID)
		r.updatingExecutionTable[id].saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads(r.reducedNT)
		r.updatingExecutionTable[id].produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads()
	}
}

func (r *Reduction) prepareUpdatingExecutionTableToBePropagatedToReducedNT() {
	if r.generativeNT != nil {
		r.mergeUpdatingExecutionTableWithUnchangedExecutionTable()
	}
}

func (r *Reduction) propagateUpdatingExecutionTableToReducedNT() {
	r.reducedNT.SetExecutionTable(&r.updatingExecutionTable)
}

func (r *Reduction) mergeUpdatingExecutionTableWithUnchangedExecutionTable() {
	unchangedExecutionTable := r.generativeNT.executionTable
	ok := r.updatingExecutionTable.merge(unchangedExecutionTable)
	if !ok {
		panic(`Reduction Handle Node error: can NOT merge execution tables`)
	}
}

// phase1 checks if the udpe's entry point matches current Reduction, without updating the path pattern,
// and, if it maches, creates new execution threads accordingly
func (r *Reduction) addNewExecutionThreadsToExecutionRecord(executionRecord *executionRecord) {
	udpe := r.globalUdpeRecordBeingConsidered.udpe()
	entryPoint := udpe.entryPoint()

	if _, _, ok := entryPoint.matchWithReductionOf(r.reducedNT.Node(), false); !ok {
		return
	}

	switch udpeType := r.globalUdpeRecordBeingConsidered.udpeType(); udpeType {
	case FPE:
		// When a node matches an FPE entry point, the thread has found a solution
		executionRecord.addExecutionThread(nil, r.reducedNT, entryPoint)
	case RPE:
		if r.wrappedNT == nil {
			return
		}
		// When a node matches an RPE entry point, the thread has found a context
		executionRecord.addExecutionThread(r.wrappedNT, nil, entryPoint)
		childrenOfWrappedNT := r.wrappedNT.Children()
		for _, child := range childrenOfWrappedNT {
			executionRecord.addExecutionThread(child, nil, udpe.entryPoint())
		}

	default:
		panic(fmt.Sprintf(`adding new execution threads to execution record: unknown udpe type %q`, udpeType))
	}
}
