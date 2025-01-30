package xpath

// TODO(vvihorev): support NUDPEs
type globalNudpeTable struct {
	list []*globalNudpeRecord
}

func (globalNudpeTable *globalNudpeTable) recordByID(id int) *globalNudpeRecord {
	return globalNudpeTable.list[id]
}

func (globalNudpeTable *globalNudpeTable) mainQueryRecord() *globalNudpeRecord {
	return globalNudpeTable.list[globalNudpeTable.size()-1]
}

func (globalNudpeTable *globalNudpeTable) addNudpeRecord(length int) *globalNudpeRecord {
	ctxSols := make(contextSolutionsMap)
	newNudpeRecord := &globalNudpeRecord{
		len:     length,
		ctxSols: ctxSols,
	}
	globalNudpeTable.list = append(globalNudpeTable.list, newNudpeRecord)
	return newNudpeRecord
}

// size returns the number of NUDPE which are recorded inside the table
func (globalNudpeTable *globalNudpeTable) size() int {
	return len(globalNudpeTable.list)
}

type globalNudpeRecord struct {
	ctxSols contextSolutionsMap
	len     int
}

func (globalNudpeRecord *globalNudpeRecord) contextSolutions() contextSolutionsMap {
	return globalNudpeRecord.ctxSols
}

func (globalNudpeRecord *globalNudpeRecord) setContextSolutions(ctxSols contextSolutionsMap) {
	globalNudpeRecord.ctxSols = ctxSols
}

// length returns the number of UDPE by which the NUDPE is composed
func (globalNudpeRecord *globalNudpeRecord) length() int {
	return globalNudpeRecord.len
}

func (globalNudpeRecord *globalNudpeRecord) hasSolutionsFor(ctx *NonTerminal) bool {
	return globalNudpeRecord.ctxSols.hasSolutionsFor(ctx)
}
