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

	// iterate over all global udpe records and execute main phases
	for id, gr := range udpeGlobalTable.list {
		r.globalUdpeRecordBeingConsidered = &gr

		// Phase 1 checks if the udpe's entry point matches current
		// Reduction, without updating the path pattern, and, if it
		// maches, creates new execution threads accordingly
		{
			executionRecord := &r.updatingExecutionTable[id]
			udpe := r.globalUdpeRecordBeingConsidered.udpe()
			entryPoint := udpe.entryPoint()

			if _, _, ok := entryPoint.matchWithReductionOf(r.reducedNT.Node(), false); ok {
				switch udpeType := r.globalUdpeRecordBeingConsidered.udpeType(); udpeType {
				case FPE:
					// When a node matches an FPE entry point, the thread has found a solution
					executionRecord.addExecutionThread(nil, r.reducedNT, entryPoint)
				case RPE:
					if r.wrappedNT != nil {
						// When a node matches an RPE entry point, the thread has found a context
						executionRecord.addExecutionThread(r.wrappedNT, nil, entryPoint)
						childrenOfWrappedNT := r.wrappedNT.Children()
						for _, child := range childrenOfWrappedNT {
							executionRecord.addExecutionThread(child, nil, udpe.entryPoint())
						}
					}
				default:
					panic(fmt.Sprintf(`adding new execution threads to execution record: unknown udpe type %q`, udpeType))
				}
			}
		}

		er := r.updatingExecutionTable[id]

		// Phase 2 - Update all execution threads
		{
			for i := 0; i < er.threads.size; i++ {
				etPathPattern := er.threads.array[i].pp

				//The path pattern of the execution thread may be empty if the thread is speculative
				//and it's not completed because of some unchecked speculation
				if etPathPattern.isEmpty() {
					continue
				}

				predicate, newPathPattern, ok := etPathPattern.matchWithReductionOf(r.reducedNT.Node(), true)
				if ok {
					var etReceivingSpeculation = er.threads.array[i]
					if newPathPattern != nil {
						etReceivingSpeculation = er.addExecutionThread(er.threads.array[i].ctx, er.threads.array[i].sol, newPathPattern)
						er.threads.array[i].addChild(etReceivingSpeculation)
					}

					if predicate != nil {
						_ = etReceivingSpeculation.addSpeculation(predicate, *r.reducedNT)
					}
				} else {
					er.removeExecutionThread(er.threads.array[i], false)
				}
			}
		}

		// stopUnfoundedSpeculativeExecutionThreads iterates over all the running execution threads and, for each speculative
		// execution thread, the speculation is evaluated. If the speculation ends up to be unfounded, the speculative execution thread,
		// and all its Children recursively, are stopped
		{
			for i := 0; i < er.threads.size; i++ {
				if areSpeculationsFounded := er.threads.array[i].checkAndUpdateSpeculations(r.updatingExecutionTable.evaluateID); !areSpeculationsFounded {
					er.removeExecutionThread(er.threads.array[i], true)
				}
			}
		}

		// saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads saves the input NonTerminal as either context or solution
		// for all the execution threads that are completed. By the time at which the execution thread is completed, it might not be able
		// to produce context-solution items because of running speculations. Even if the execution thread can not produce context-solution
		// items, it has to save the non terminal whose reducetion caused the execution thread to complete
		{
			for i := 0; i < er.threads.size; i++ {
				if er.threads.array[i].pp.isEmpty() {
					if er.threads.array[i].ctx == nil {
						er.threads.array[i].ctx = r.reducedNT
						break
					}

					if er.threads.array[i].sol == nil {
						er.threads.array[i].sol = r.reducedNT
					}
				}
			}
		}

		// produce context solutions out of completed non speculative execution threads
		{
			for i := 0; i < er.threads.size; i++ {
				if er.threads.array[i].pp.isEmpty() && !er.threads.array[i].isSpeculative() {
					er.ctxSols.addContextSolution(er.threads.array[i].ctx, er.threads.array[i].sol)
					er.removeExecutionThread(er.threads.array[i], false)
				}
			}
		}
	}

	if r.generativeNT != nil {
		// merge updating execution table with unchanged execution table
		unchangedExecutionTable := r.generativeNT.executionTable
		ok := r.updatingExecutionTable.merge(unchangedExecutionTable)
		if !ok {
			panic(`Reduction Handle Node error: can NOT merge execution tables`)
		}
	}
	// propagate updating execution table to reduced NT
	r.reducedNT.SetExecutionTable(&r.updatingExecutionTable)
}

