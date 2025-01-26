package xpath

import (
	"fmt"
)

type executionTableIterableCallback func(id int, er executionRecord) (doBreak bool)

// TODO(vvihorev): maybe store the reference to global NUDPE in the table, instead
// of duplicating it to all records in the table.
// TODO(vvihorev): if we only access records through the table, records dont need
// to teep a pointer to their table.
type executionTable struct {
	list []executionRecord
}

func (et *executionTable) actualList() []executionRecord {
	return et.list
}

func (et *executionTable) recordByID(id int) (execRecord *executionRecord, err error) {
	defer func() {
		if r := recover(); r != nil {
			execRecord = nil
			err = fmt.Errorf("execution table lookup: can NOT get execution record with id %d", id)
			return
		}
	}()
	execRecord = &et.list[id]
	return
}

func (et *executionTable) mainQueryRecord() executionRecord {
	return et.list[et.size()-1]
}

// merge joins the incoming execution table to the receiving execution table and returns
// the receiving execution table
func (et *executionTable) merge(incoming *executionTable) (result *executionTable, ok bool) {
	result = et
	ok = true

	for id, er := range et.list {
		incomingRecord, err := incoming.recordByID(id)
		if err != nil {
			ok = false
			break
		}

		if _, isMerged := er.merge(incomingRecord); !isMerged {
			ok = false
			break
		}
	}
	return
}

func (et *executionTable) size() int {
	return len(et.list)
}

// evaluateID returns the boolean value of an udpe with a certain id w.r.t a specific context.
func (et *executionTable) evaluateID(udpeID int, context NonTerminal, evaluationsCount int) customBool {
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
	if record.udpeType() == FPE {
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
	expType      udpeType
	t            *executionTable
	ctxSols      contextSolutionsMap
	threads      swapbackArray[executionThread]
	gNudpeRecord globalNudpeRecord
}

func (er *executionRecord) String() string {
	return fmt.Sprintf("{ %v | %v | %v }", er.expType, er.ctxSols, er.threads)
}

func (er *executionRecord) addExecutionThread(ctx, sol NonTerminal, pp pathPattern) (et executionThread) {
	et = executionThread{
		ctx:    ctx,
		sol:    sol,
		pp:     pp,
		speculations: swapbackArray[speculation]{
			array: make([]speculation, 0),
		},
	}
	er.threads.append(et)
	logger.Printf("adding execution thread: %v", et)
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

func (er *executionRecord) hasExecutionThreadRunningFor(ctx NonTerminal) bool {
	for i := 0; i < er.threads.size; i++ {
		found := er.threads.array[i].ctx == ctx
		if found {
			return true
		}
	}
	return false
}

func (er *executionRecord) hasSolutionsFor(ctx NonTerminal) bool {
	return er.ctxSols.hasSolutionsFor(ctx)
}

func (er *executionRecord) contextSolutions() contextSolutionsMap {
	return er.ctxSols
}

func (er *executionRecord) udpeType() udpeType {
	return er.expType
}

func (er *executionRecord) nudpeRecord() globalNudpeRecord {
	return er.gNudpeRecord
}

func (er *executionRecord) belongsToNudpe() bool {
	return er.gNudpeRecord != nil
}

// updateExecutionThreads takes the node being reduced and asks all the running execution threads
// to update accordingly
func (er *executionRecord) updateAllExecutionThreads(reduced NonTerminal) {
	for _, et := range er.threads.array {
		etPathPattern := et.pp
		//The path pattern of the execution thread may be empty if the thread is speculative
		//and it's not completed because of some unchecked speculation
		if !etPathPattern.isEmpty() {
			etReprBeforeUpdate := fmt.Sprintf("%v", et)
			predicate, newPathPattern, ok := etPathPattern.matchWithReductionOf(reduced.Node(), true)
			if ok {
				etReprAfterUpdate := fmt.Sprintf("%v", et)
				logger.Printf("updated execution thread: %s -> %s", etReprBeforeUpdate, etReprAfterUpdate)

				var etReceivingSpeculation = et
				if newPathPattern != nil {
					etReceivingSpeculation = er.addExecutionThread(et.ctx, et.sol, newPathPattern)
					et.addChild(etReceivingSpeculation)
				}

				if predicate != nil {
					sp := etReceivingSpeculation.addSpeculation(predicate, reduced)
					logger.Printf("adding speculation: %v to execution thread %v", sp, et)
				}
			} else {
				logger.Printf("removing execution thread beacuse path pattern does NOT match: %s", etReprBeforeUpdate)
				er.removeExecutionThread(et, false)
			}
		}
	}
}

// stopUnfoundedSpeculativeExecutionThreads iterates over all the running execution threads and, for each speculative
// execution thread, the speculation is evaluated. If the speculation ends up to be unfounded, the speculative execution thread,
// and all its Children recursively, are stopped
func (er *executionRecord) stopUnfoundedSpeculativeExecutionThreads(evaluator evaluator) {
	for _, et := range er.threads.array {
		if areSpeculationsFounded := et.checkAndUpdateSpeculations(evaluator); !areSpeculationsFounded {
			er.removeExecutionThread(et, true)
		}
	}
}

// saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads saves the input NonTerminal as either context or solution
// for all the execution threads that are completed. By the time at which the execution thread is completed, it might not be able
// to produce context-solution items because of running speculations. Even if the execution thread can not produce context-solution
// items, it has to save the non terminal whose reducetion caused the execution thread to complete
func (er *executionRecord) saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads(contextOrSolution NonTerminal) {
	for _, et := range er.threads.array {
		if et.pp.isEmpty() {
			if et.ctx == nil {
				et.ctx = contextOrSolution
				return
			}

			if et.sol == nil {
				et.sol = contextOrSolution
			}
		}
	}
}

// produceContextSolutions produce new context solutions from completed execution threads and removes
// completed execution threads
func (er *executionRecord) produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads() {
	for _, et := range er.threads.array {
		if et.pp.isEmpty() && !et.isSpeculative() {
			logger.Printf("adding context-solution: (%v , %v)", et.ctx, et.sol)
			er.ctxSols.addContextSolution(et.ctx, et.sol)
			er.removeExecutionThread(et, false)
		}
	}
}

func (er *executionRecord) merge(incoming *executionRecord) (result executionRecord, ok bool) {
	result = *er

	if incoming != nil {
		for i := 0; i < incoming.threads.size; i++ {
			er.threads.append(incoming.threads.array[i])
		}
	}
	_, ok = er.ctxSols.merge(incoming.ctxSols)
	return
}
