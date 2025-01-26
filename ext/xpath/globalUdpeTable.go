package xpath

type globalTableIterableCallback func(id int, record globalUdpeRecord)

type globalUdpeTable struct {
	list []globalUdpeRecord
}

func (globalUdpeTable *globalUdpeTable) newExecutionTable() *executionTable {
	et := new(executionTable)
	globalUdpeTableSize := globalUdpeTable.size()
	executionRecordsGroup := make([]executionRecord, globalUdpeTableSize)
	for id := range executionRecordsGroup {
		globalUdpeRecord := globalUdpeTable.recordByID(id)
		executionRecordsGroup[id] = executionRecord{
			expType:      globalUdpeRecord.udpeType(),
			t:            et,
			ctxSols:      newContextSolutionsMap(),
			threads:       swapbackArray[executionThread]{
				array: make([]executionThread, 0),
			},
			gNudpeRecord: globalUdpeRecord.nudpeRecord(),
		}
	}
	et.list = executionRecordsGroup

	return et
}

func (globalUdpeTable *globalUdpeTable) actualList() []globalUdpeRecord {
	return globalUdpeTable.list
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

func (globalUdpeTable *globalUdpeTable) addFpe(fpe fpe) (id int, record globalUdpeRecord) {
	return globalUdpeTable.addUdpe(fpe, FPE)
}

func (globalUdpeTable *globalUdpeTable) addRpe(rpe rpe) (id int, record globalUdpeRecord) {
	return globalUdpeTable.addUdpe(rpe, RPE)
}

// addUdpe creates a new record inside the global udpe table
func (globalUdpeTable *globalUdpeTable) addUdpe(udpe udpe, udpeType udpeType) (id int, record globalUdpeRecord) {
	r := &globalUdpeRecordImpl{
		exp:     udpe,
		expType: udpeType,
	}
	globalUdpeTable.list = append(globalUdpeTable.list, r)
	return len(globalUdpeTable.list) - 1, r
}

type globalUdpeRecord interface {
	udpe() udpe
	udpeType() udpeType
	nudpeRecord() globalNudpeRecord
	setNudpeRecord(nudpeRecord globalNudpeRecord)
}

type globalUdpeRecordImpl struct {
	exp          udpe
	expType      udpeType
	gNudpeRecord globalNudpeRecord
}

func (globalUdpeRecord *globalUdpeRecordImpl) udpe() udpe {
	return globalUdpeRecord.exp

}

// udpeType returns the type of the underlying UDPE
func (globalUdpeRecord *globalUdpeRecordImpl) udpeType() udpeType {
	return globalUdpeRecord.expType
}

// nudpeRecord returns the globalNudpeRecord of the NUDPE to which the UDPE belongs
func (globalUdpeRecord *globalUdpeRecordImpl) nudpeRecord() globalNudpeRecord {
	return globalUdpeRecord.gNudpeRecord
}

// TODO(vvihorev): does the XPath evaluator work with a single topmost NUDPE,
// or the general case is that a NUDPE can have other NUDPEs in predicates?
// Do we really need to point all UDPEs back to the single NUDPE?
func (globalUdpeRecord *globalUdpeRecordImpl) setNudpeRecord(nudpeRecord globalNudpeRecord) {
	globalUdpeRecord.gNudpeRecord = nudpeRecord
}
