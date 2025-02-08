package xpath

import (
	"fmt"
)

type Reduction struct {
	reducedNT, generativeNT, wrappedNT *NonTerminal
	updatingExecutionTable             executionTable
	globalUdpeRecordBeingConsidered    *globalUdpeRecord
}

// TODO(vvihorev): further simplify the call to reduction
func (r *Reduction) Setup(reducedNT, generativeNT, wrappedNT *NonTerminal) {
	var updatingExecutionTable executionTable
	if wrappedNT != nil {
		updatingExecutionTable = wrappedNT.executionTable
	} else {
		updatingExecutionTable = udpeGlobalTable.newExecutionTable()
	}

	if DEBUG {
		logger.Printf("REDUCTION: %v <-- %v <> %v </>", reducedNT.String(), generativeNT.String(), wrappedNT.String())
	}

	r.reducedNT = reducedNT
	r.generativeNT = generativeNT
	r.wrappedNT = wrappedNT
	r.updatingExecutionTable = updatingExecutionTable

	// iterate over all global udpe records and execute main phases
	for id, gr := range udpeGlobalTable.list {
		r.globalUdpeRecordBeingConsidered = &gr
		if DEBUG {
			logger.Printf(" UDPE %d", id)
		}

		// Phase 1 checks if the udpe's entry point matches current
		// Reduction, without updating the path pattern, and, if it
		// maches, creates new execution threads accordingly
		{
			if DEBUG {
				logger.Printf("  PHASE 1")
			}
			udpe := r.globalUdpeRecordBeingConsidered.udpe()
			entryPoint := udpe.entryPoint()

			if _, _, ok := entryPoint.matchWithReductionOf(r.reducedNT.Node(), false); ok {
				switch udpeType := r.globalUdpeRecordBeingConsidered.udpeType(); udpeType {
				case FPE:
					// When a node matches an FPE entry point, the thread has found a solution
					if DEBUG {
						et := r.updatingExecutionTable.records[id].addExecutionThread(nil, r.reducedNT, entryPoint)
						logger.Printf("   adding execution thread: %v", et.String())
					} else {
						r.updatingExecutionTable.records[id].addExecutionThread(nil, r.reducedNT, entryPoint)
					}
				case RPE:
					if r.wrappedNT != nil {
						// When a node matches an RPE entry point, the thread has found a context
						if DEBUG {
							et := r.updatingExecutionTable.records[id].addExecutionThread(r.wrappedNT, nil, entryPoint)
							logger.Printf("   adding execution thread: %v", et.String())
						} else {
							r.updatingExecutionTable.records[id].addExecutionThread(r.wrappedNT, nil, entryPoint)
						}
						childrenOfWrappedNT := r.wrappedNT.Children()
						for _, child := range childrenOfWrappedNT {
							r.updatingExecutionTable.records[id].addExecutionThread(child, nil, udpe.entryPoint())
						}
					}
				default:
					panic(fmt.Sprintf(`adding new execution threads to execution record: unknown udpe type %q`, udpeType))
				}
			}
		}

		// Phase 2 - Update all execution threads
		{
			var predicate *predicate
			var newPathPattern pathPattern
			var ok bool

			if DEBUG {
				logger.Printf("  PHASE 2")
			}
			for i := 0; i < r.updatingExecutionTable.records[id].threads.size; i++ {
				etPathPattern := r.updatingExecutionTable.records[id].threads.array[i].pp

				//The path pattern of the execution thread may be empty if the thread is speculative
				//and it's not completed because of some unchecked speculation
				if etPathPattern.isEmpty() {
					continue
				}

				if DEBUG {
					ppBefore := etPathPattern.String()
					predicate, newPathPattern, ok = etPathPattern.matchWithReductionOf(r.reducedNT.Node(), true)
					logger.Printf("   updating execution thread: %v -> %v", ppBefore, etPathPattern.String())
				} else {
					predicate, newPathPattern, ok = etPathPattern.matchWithReductionOf(r.reducedNT.Node(), true)
				}

				if ok {
					if newPathPattern != nil {
						etReceivingSpeculation := r.updatingExecutionTable.records[id].addExecutionThread(r.updatingExecutionTable.records[id].threads.array[i].ctx, r.updatingExecutionTable.records[id].threads.array[i].sol, newPathPattern)
						r.updatingExecutionTable.records[id].threads.array[i].addChild(etReceivingSpeculation)
						r.updatingExecutionTable.records[id].threads.array[i].pp = newPathPattern
						if predicate != nil {
							etReceivingSpeculation.addSpeculation(predicate, *r.reducedNT)
						}
					} else {
						if predicate != nil {
							r.updatingExecutionTable.records[id].threads.array[i].addSpeculation(predicate, *r.reducedNT)
						}
					}

				} else {
					r.updatingExecutionTable.records[id].removeExecutionThread(r.updatingExecutionTable.records[id].threads.array[i], false)
				}
			}
		}

		// Phase 3 - Stop speculative threads, store context-solution pairs for compeleted threads
		{
			if DEBUG {
				logger.Printf("  PHASE 3")
			}
			for i := 0; i < r.updatingExecutionTable.records[id].threads.size; i++ {
				// If a thread speculation is unfounded, stop the speculative execution thread, and all its Children recursively
				if areSpeculationsFounded := r.updatingExecutionTable.records[id].threads.array[i].checkAndUpdateSpeculations(r.updatingExecutionTable.evaluateID); !areSpeculationsFounded {
					r.updatingExecutionTable.records[id].removeExecutionThread(r.updatingExecutionTable.records[id].threads.array[i], true)
				}
			}

			for i := 0; i < r.updatingExecutionTable.records[id].threads.size; i++ {
				// for all completed threads save the input NonTerminal as either context or solution.
				// By the time at which the execution thread is completed, it might not be able
				// to produce context-solution items because of running speculations.
				// Even if the execution thread can not produce context-solution items, it has to save the non
				// terminal whose reducetion caused the execution thread to complete
				if r.updatingExecutionTable.records[id].threads.array[i].pp.isEmpty() {
					if r.updatingExecutionTable.records[id].threads.array[i].ctx == nil {
						r.updatingExecutionTable.records[id].threads.array[i].ctx = r.reducedNT
					} else if r.updatingExecutionTable.records[id].threads.array[i].sol == nil {
						r.updatingExecutionTable.records[id].threads.array[i].sol = r.reducedNT
					}
				}
			}

			for i := 0; i < r.updatingExecutionTable.records[id].threads.size; i++ {
				// produce context solutions out of completed non speculative execution threads
				if DEBUG {
					logger.Printf("   checking context-solutions for: %v", r.updatingExecutionTable.records[id].threads.array[i].String())
				}

				if r.updatingExecutionTable.records[id].threads.array[i].pp.isEmpty() && !r.updatingExecutionTable.records[id].threads.array[i].isSpeculative() {
					r.updatingExecutionTable.records[id].ctxSols.addContextSolution(r.updatingExecutionTable.records[id].threads.array[i].ctx.Position().Start(), r.updatingExecutionTable.records[id].threads.array[i].ctx.Position().End(), r.updatingExecutionTable.records[id].threads.array[i].sol)
					if DEBUG {
						logger.Printf("   added a context-solution: [ %v, %v ] to record: %d %v", r.updatingExecutionTable.records[id].threads.array[i].ctx.String(), r.updatingExecutionTable.records[id].threads.array[i].sol.String(), i, r.updatingExecutionTable.records[id].String())
					}
					r.updatingExecutionTable.records[id].removeExecutionThread(r.updatingExecutionTable.records[id].threads.array[i], false)
				}
			}
		}
	}

	if DEBUG {
		logger.Printf(" MERGE")
	}

	if r.generativeNT != nil {
		// merge updating execution table with unchanged execution table
		unchangedExecutionTable := r.generativeNT.executionTable
		ok := r.updatingExecutionTable.merge(&unchangedExecutionTable)
		if !ok {
			panic(`Reduction Handle Node error: can NOT merge execution tables`)
		}
	}

	if DEBUG {
		for i, er := range updatingExecutionTable.records {
			if r.generativeNT == nil {
				logger.Printf("  skipping merge %d %v", i, er.String())
			} else {
				logger.Printf("  after merge %d %v", i, er.String())
			}

			for i := 0; i < er.threads.size; i++ {
				logger.Printf("  own thread: %v", er.threads.array[i].String())
			}
		}
	}
	// propagate updating execution table to reduced NT
	r.reducedNT.executionTable = r.updatingExecutionTable
}
