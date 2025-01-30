package xpath

import (
	"fmt"
)

type executionTableIterableCallback func(id int, er executionRecord) (doBreak bool)

// TODO(vvihorev): don't pass the pointer to the table (slice)
type executionTable []executionRecord

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

		if incoming != nil {
			for i := 0; i < incomingRecord.threads.size; i++ {
				er.threads.append(incomingRecord.threads.array[i])
			}
		}
		er.ctxSols.merge(incomingRecord.ctxSols)
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
	if record.gNudpeRecord != nil {
		return Undefined
	}

	if record.ctxSols.hasSolutionsFor(context) {
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
