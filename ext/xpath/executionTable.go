package xpath

import (
	"fmt"
	"strings"
)

type executionTableIterableCallback func(id int, er executionRecord) (doBreak bool)

type executionTable []executionRecord

// NOTE(vvihorev): Use for debugging only
func (et *executionTable) String() string {
	records := make([]string, 0)
	for _, er := range *et {
		records = append(records, er.String())
	}
	return strings.Join(records, ",\n")
}

func (et executionTable) recordByID(id int) (execRecord *executionRecord, err error) {
	defer func() {
		if r := recover(); r != nil {
			execRecord = nil
			err = fmt.Errorf("execution table lookup: can NOT get execution record with id %d", id)
			return
		}
	}()
	execRecord = &et[id]
	return
}

func (et executionTable) mainQueryRecord() executionRecord {
	return et[len(et)-1]
}

// merge joins the incoming execution table to the receiving execution table and returns
// the receiving execution table
func (et executionTable) merge(incoming *executionTable) (ok bool) {
	ok = true

	for id, er := range et {
		incomingRecord, err := incoming.recordByID(id)
		if err != nil {
			ok = false
			break
		}

		er.merge(incomingRecord)
	}
	return
}

// evaluateID returns the boolean value of an udpe with a certain id w.r.t a specific context.
func (et *executionTable) evaluateID(udpeID int, context *NonTerminal, evaluationsCount int) customBool {
	record, err := et.recordByID(udpeID)
	if err != nil {
		panic(fmt.Sprintf(`udpe boolean evaluation error for id: %v`, err))
	}

	//Do not consider the effective boolean value of all those UDPEs which are
	//the initial UDPE of some NUDPE.
	//The boolean value of a NUDPE is computed at the end of the entire parsing of the document.
	if record.belongsToNudpe() {
		return Undefined
	}

	if record.hasSolutionsFor(context) {
		return True
	}

	//FPEs are always bounded w.r.t. a specific context and their boolean value
	//is always available as soon as the context has been synthesized.
	if record.expType == FPE {
		return False
	}

	//RPEs may be unbounded w.r.t. a certain context and their boolean values
	//may be undefined at the time when the context is synthesized.
	if evaluationsCount == 0 || record.hasExecutionThreadRunningFor(context) {
		return Undefined
	}

	return False
}

type executionRecord struct {
	threads      swapbackArray[executionThread]
	ctxSols      contextSolutionsMap
	gNudpeRecord *globalNudpeRecord
	expType      udpeType
}

func (er *executionRecord) String() string {
	return fmt.Sprintf("{ %v | %v | %v }", er.expType, er.ctxSols, er.threads)
}

func (er *executionRecord) addExecutionThread(ctx, sol *NonTerminal, pp pathPattern) (et executionThread) {
	et = executionThread{
		ctx:    ctx,
		sol:    sol,
		pp:     pp,
		speculations: swapbackArray[speculation]{
			array: make([]speculation, 0),
		},
	}
	er.threads.append(et)
	// logger.Printf("adding execution thread: %v", et.String())
	return
}

func (er *executionRecord) removeExecutionThread(et executionThread, removeChildren bool) {
	if removeChildren {
		for _, childEt := range et.offspr {
			er.removeExecutionThread(childEt, true)
		}
	}
	er.threads.remove(et)
}

func (er *executionRecord) hasExecutionThreadRunningFor(ctx *NonTerminal) bool {
	for i := 0; i < er.threads.size; i++ {
		found := er.threads.array[i].ctx == ctx
		if found {
			return true
		}
	}
	return false
}

func (er *executionRecord) hasSolutionsFor(ctx *NonTerminal) bool {
	return er.ctxSols.hasSolutionsFor(ctx)
}

func (er *executionRecord) belongsToNudpe() bool {
	return er.gNudpeRecord != nil
}

// updateExecutionThreads takes the node being reduced and asks all the running execution threads
// to update accordingly
func (er *executionRecord) updateAllExecutionThreads(reduced *NonTerminal) {
	for i := 0; i < er.threads.size; i++ {
		etPathPattern := er.threads.array[i].pp
		//The path pattern of the execution thread may be empty if the thread is speculative
		//and it's not completed because of some unchecked speculation
		if !etPathPattern.isEmpty() {
			// etReprBeforeUpdate := fmt.Sprintf("%v", er.threads.array[i].String())
			predicate, newPathPattern, ok := etPathPattern.matchWithReductionOf(reduced.Node(), true)
			if ok {
				// etReprAfterUpdate := fmt.Sprintf("%v", er.threads.array[i].String())
				// logger.Printf("updated execution thread: %s -> %s", etReprBeforeUpdate, etReprAfterUpdate)

				var etReceivingSpeculation = er.threads.array[i]
				if newPathPattern != nil {
					etReceivingSpeculation = er.addExecutionThread(er.threads.array[i].ctx, er.threads.array[i].sol, newPathPattern)
					er.threads.array[i].addChild(etReceivingSpeculation)
				}

				if predicate != nil {
					_ = etReceivingSpeculation.addSpeculation(predicate, *reduced)
					// logger.Printf("adding speculation: %v to execution thread %v", sp, er.threads.array[i])
				}
			} else {
				// logger.Printf("removing execution thread beacuse path pattern does NOT match: %s", etReprBeforeUpdate)
				er.removeExecutionThread(er.threads.array[i], false)
			}
		}
	}
}

// stopUnfoundedSpeculativeExecutionThreads iterates over all the running execution threads and, for each speculative
// execution thread, the speculation is evaluated. If the speculation ends up to be unfounded, the speculative execution thread,
// and all its Children recursively, are stopped
func (er *executionRecord) stopUnfoundedSpeculativeExecutionThreads(evaluator evaluator) {
	for i := 0; i < er.threads.size; i++ {
		if areSpeculationsFounded := er.threads.array[i].checkAndUpdateSpeculations(evaluator); !areSpeculationsFounded {
			er.removeExecutionThread(er.threads.array[i], true)
		}
	}
}

// saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads saves the input NonTerminal as either context or solution
// for all the execution threads that are completed. By the time at which the execution thread is completed, it might not be able
// to produce context-solution items because of running speculations. Even if the execution thread can not produce context-solution
// items, it has to save the non terminal whose reducetion caused the execution thread to complete
func (er *executionRecord) saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads(contextOrSolution *NonTerminal) {
	for i := 0; i < er.threads.size; i++ {
		if er.threads.array[i].pp.isEmpty() {
			if er.threads.array[i].ctx == nil {
				er.threads.array[i].ctx = contextOrSolution
				return
			}

			if er.threads.array[i].sol == nil {
				er.threads.array[i].sol = contextOrSolution
			}
		}
	}
}

// produceContextSolutions produce new context solutions from completed execution threads and removes
// completed execution threads
func (er *executionRecord) produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads() {
	for i := 0; i < er.threads.size; i++ {
		if er.threads.array[i].pp.isEmpty() && !er.threads.array[i].isSpeculative() {
			// logger.Printf("adding context-solution: (%v , %v)", er.threads.array[i].ctx, er.threads.array[i].sol)
			er.ctxSols.addContextSolution(er.threads.array[i].ctx, er.threads.array[i].sol)
			er.removeExecutionThread(er.threads.array[i], false)
		}
	}
}

func (er *executionRecord) merge(incoming *executionRecord) {
	if incoming != nil {
		for i := 0; i < incoming.threads.size; i++ {
			er.threads.append(incoming.threads.array[i])
		}
	}
	er.ctxSols.merge(incoming.ctxSols)
}
