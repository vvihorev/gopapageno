package xpath

import (
	"fmt"
)

type Reduction struct {
	reducedNT, generativeNT, wrappedNT *NonTerminal
	updatingExecutionTable             *executionTable
	globalUdpeRecordBeingConsidered    *globalUdpeRecord
}

func (r *Reduction) Setup(reducedNT, generativeNT, wrappedNT *NonTerminal) {
	logger.Printf("Reduction.Setup: %v <= %v <> %v </>", reducedNT.String(), generativeNT.String(), wrappedNT.String())
	var updatingExecutionTable executionTable

	if wrappedNT != nil {
		updatingExecutionTable = *wrappedNT.executionTable
	} else {
		updatingExecutionTable = udpeGlobalTable.newExecutionTable()
	}

	r.reducedNT = reducedNT
	r.generativeNT = generativeNT
	r.wrappedNT = wrappedNT
	r.updatingExecutionTable = &updatingExecutionTable
}

func (r *Reduction) Handle() {
	r.iterateOverAllGlobalUdpeRecordsAndExecuteMainPhases()
	r.prepareUpdatingExecutionTableToBePropagatedToReducedNT()
	r.propagateUpdatingExecutionTableToReducedNT()
}

func (r *Reduction) iterateOverAllGlobalUdpeRecordsAndExecuteMainPhases() {
	// TODO(vvihorev): stop dereferncing slices?
	tableRecords := *r.updatingExecutionTable

	for id, gr := range udpeGlobalTable.list {
		r.globalUdpeRecordBeingConsidered = &gr

		if id < 0 || id >= len(tableRecords) {
			panic(fmt.Sprintf("cannot retrieve execution record for udpe with id: %d", id))
		}
		updatingExecutionRecord := tableRecords[id]
		logger.Printf("Considering UDPE: %v", gr.exp.entryPoint())

		r.addNewExecutionThreadsToExecutionRecord(&updatingExecutionRecord)

		updatingExecutionRecord.updateAllExecutionThreads(r.reducedNT)
		updatingExecutionRecord.stopUnfoundedSpeculativeExecutionThreads(r.updatingExecutionTable.evaluateID)
		updatingExecutionRecord.saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads(r.reducedNT)
		updatingExecutionRecord.produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads()
	}
}

func (r *Reduction) prepareUpdatingExecutionTableToBePropagatedToReducedNT() {
	if r.generativeNT != nil {
		r.mergeUpdatingExecutionTableWithUnchangedExecutionTable()
	}
}

func (r *Reduction) propagateUpdatingExecutionTableToReducedNT() {
	r.reducedNT.SetExecutionTable(r.updatingExecutionTable)
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
		executionRecord.addExecutionThread(nil, r.reducedNT, entryPoint)
	case RPE:
		if r.wrappedNT == nil {
			return
		}
		// TODO(vvihorev): what and why?
		executionRecord.addExecutionThread(r.wrappedNT, nil, entryPoint)
		childrenOfWrappedNT := r.wrappedNT.Children()
		for _, child := range childrenOfWrappedNT {
			executionRecord.addExecutionThread(child, nil, udpe.entryPoint())
		}

	default:
		panic(fmt.Sprintf(`adding new execution threads to execution record: unknown udpe type %q`, udpeType))
	}
}
