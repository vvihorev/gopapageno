package xpath

type globalTableIterableCallback func(id int, record globalUdpeRecord)

type globalUdpeTable struct {
	list []globalUdpeRecord
}

func (globalUdpeTable *globalUdpeTable) newExecutionTable() executionTable {
	globalUdpeTableSize := globalUdpeTable.size()
	records := make([]executionRecord, globalUdpeTableSize)

	for id := range records {
		globalUdpeRecord := globalUdpeTable.recordByID(id)
		records[id].expType = globalUdpeRecord.udpeType()
		records[id].ctxSols = make(contextSolutionsMap)
		records[id].threads = swapbackArray[executionThread]{
			array: make([]executionThread, 0),
		}
		records[id].gNudpeRecord = globalUdpeRecord.nudpeRecord()
	}
	return executionTable{records}
}

func (globalUdpeTable *globalUdpeTable) size() int {
	return len(globalUdpeTable.list)
}

func (globalUdpeTable *globalUdpeTable) recordByID(id int) globalUdpeRecord {
	return globalUdpeTable.list[id]
}

func (globalUdpeTable *globalUdpeTable) mainQueryRecord() globalUdpeRecord {
	return globalUdpeTable.list[globalUdpeTable.size()-1]
}

func (globalUdpeTable *globalUdpeTable) addFpe(fpe *fpe) (id int, record globalUdpeRecord) {
	return globalUdpeTable.addUdpe(fpe, FPE)
}

func (globalUdpeTable *globalUdpeTable) addRpe(rpe *rpe) (id int, record globalUdpeRecord) {
	return globalUdpeTable.addUdpe(rpe, RPE)
}

// addUdpe creates a new record inside the global udpe table
func (globalUdpeTable *globalUdpeTable) addUdpe(udpe udpe, udpeType udpeType) (id int, record globalUdpeRecord) {
	r := globalUdpeRecord{
		exp:     udpe,
		expType: udpeType,
	}
	globalUdpeTable.list = append(globalUdpeTable.list, r)
	return len(globalUdpeTable.list) - 1, r
}

type globalUdpeRecord struct {
	exp          udpe
	expType      udpeType
	gNudpeRecord *globalNudpeRecord
}

func (globalUdpeRecord *globalUdpeRecord) udpe() udpe {
	return globalUdpeRecord.exp

}

// udpeType returns the type of the underlying UDPE
func (globalUdpeRecord *globalUdpeRecord) udpeType() udpeType {
	return globalUdpeRecord.expType
}

// nudpeRecord returns the globalNudpeRecord of the NUDPE to which the UDPE belongs
func (globalUdpeRecord *globalUdpeRecord) nudpeRecord() *globalNudpeRecord {
	return globalUdpeRecord.gNudpeRecord
}

func (globalUdpeRecord *globalUdpeRecord) setNudpeRecord(nudpeRecord *globalNudpeRecord) {
	globalUdpeRecord.gNudpeRecord = nudpeRecord
}
