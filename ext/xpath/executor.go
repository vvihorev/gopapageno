package xpath

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/giornetta/gopapageno"
)

const defaultExecutorNumberOfThreads = 1

// Singletons
var nudpeGlobalTable *globalNudpeTable
var udpeGlobalTable *globalUdpeTable
var logger Logger

func init() {
	logger = newNopLogger()
}

type mainQueryType int

const (
	UDPE mainQueryType = iota
	NUDPE
)

type executor struct {
	numberOfThreads         int
	mainQueryType           mainQueryType
	resultingExecutionTable *executionTable

	source []byte
	runner *gopapageno.Runner
}

// ExecutorCommand represents a command that can be made by a
// client to execute a XPath query
type ExecutorCommand struct {
	xpathQuery      string
	source          []byte
	numberOfThreads int
	verbose         bool
}

// Execute specify the XPath query to be executed
func Execute(xpathQuery string) *ExecutorCommand {
	return &ExecutorCommand{
		xpathQuery: xpathQuery,
	}
}

func (executorCommand *ExecutorCommand) WithNumberOfThreads(numberOfThreads int) *ExecutorCommand {
	executorCommand.numberOfThreads = numberOfThreads
	return executorCommand
}

func (executorCommand *ExecutorCommand) InVerboseMode() *ExecutorCommand {
	executorCommand.verbose = true
	logger = log.New(os.Stderr, "", log.LstdFlags)
	return executorCommand
}

func (executorCommand *ExecutorCommand) Against(source []byte) *ExecutorCommand {
	executorCommand.source = source
	return executorCommand
}

// Run takes care of executing the command and to return
func (executorCommand *ExecutorCommand) Run(runner *gopapageno.Runner) (results []Position, err error) {
	executor := &executor{
		numberOfThreads: defaultExecutorNumberOfThreads,
		runner:          runner,
	}

	executor.numberOfThreads = executorCommand.numberOfThreads
	executor.source = executorCommand.source

	executor.parseQuery(executorCommand.xpathQuery)

	err = executor.executeUDPEsWhileParsing()
	if err != nil {
		return
	}
	
	err = executor.completeExecutionOfUDPEsAndNUDPEs()
	if err != nil {
		return
	}

	results = executor.retrieveResults()

	return
}

func (executor executor) parseQuery(xpathQuery string) {
	udpeGlobalTable = new(globalUdpeTable)
	nudpeGlobalTable = new(globalNudpeTable)

	i := 0
	peek := func() byte {
		if i+1 >= len(xpathQuery) {
			return 0
		}
		return xpathQuery[i+1]
	}

	readTag := func() string {
		start := i
		for i < len(xpathQuery) && xpathQuery[i] != '[' && xpathQuery[i] != '/' && xpathQuery[i] != '\\' {
			i++
		}
		if i == start+1 {
			panic("malformed query: expected to find a tag name in Test")
		}
		return xpathQuery[start:i]
	}

	var curFpe *fpeBuilder
	var curRpe *rpeBuilder
	hasFpe := false
	hasRpe := false

	// TODO(vvihorev): support predicates
	// p := &predicate{
	// 	expressionVector: []operator{and(), atom(), atom()},
	// 	atomsLookup: map[atomID]int{
	// 		atomID(fpe1ID): 1,
	// 		atomID(fpe2ID): 2,
	// 	},
	// }
	// TODO(vvihorev): support attributes
	// TODO(vvihorev): support text builtin
	for i < len(xpathQuery) {
		c := xpathQuery[i]

		switch c {
		case '/':
			hasFpe = true
			if curFpe == nil {
				curFpe = &fpeBuilder{}
			}
			if curRpe != nil {
				udpeGlobalTable.addRpe(curRpe.end())
				curRpe = nil
			}

			if peek() == '/' {
				curFpe.addAxis(descendantOrSelf)
				i += 2
			} else if peek() == 0 {
				panic("malformed query: expected to find Test after forward Axis")
			} else {
				curFpe.addAxis(child)
				i += 1
			}

			curFpe.addUdpeTest(newElementTest(readTag(), nil, nil))

		case '\\':
			hasRpe = true
			if curRpe == nil {
				curRpe = &rpeBuilder{}
			}
			if curFpe != nil {
				udpeGlobalTable.addFpe(curFpe.end())
				curFpe = nil
			}

			if peek() == '\\' {
				curRpe.addAxis(ancestorOrSelf)
				i += 2
			} else if peek() == 0 {
				panic("malformed query: expected to find Test after reverse Axis")
			} else {
				curRpe.addAxis(parent)
				i += 1
			}

			curRpe.addUdpeTest(newElementTest(readTag(), nil, nil))

		default:
			panic("malformed query: expected an Axis")
		}
	}

	if curFpe != nil {
		udpeGlobalTable.addFpe(curFpe.end())
	}
	if curRpe != nil {
		udpeGlobalTable.addRpe(curRpe.end())
	}

	if hasFpe && hasRpe {
		executor.mainQueryType = NUDPE
	} else {
		executor.mainQueryType = UDPE
	}
}

func (executor *executor) executeUDPEsWhileParsing() error {
	axiomToken, err := executor.runner.Run(context.Background(), executor.source)
	if err != nil {
		return fmt.Errorf("could not parse: %v", err)
	}
	executor.resultingExecutionTable = axiomToken.Value.(*NonTerminal).executionTable

	return err
}

func (executor *executor) completeExecutionOfUDPEsAndNUDPEs() (err error) {

	if numberOfNUDPEs := nudpeGlobalTable.size(); numberOfNUDPEs == 0 {
		return
	}

	var currentNudpeRecord *globalNudpeRecord
	var currentNudpeContextSolutionsMaps []contextSolutionsMap

	for _, er := range *executor.resultingExecutionTable {
		if er.belongsToNudpe() {
			if er.gNudpeRecord != currentNudpeRecord {
				if currentNudpeRecord != nil {
					currentNudpeRecord.setContextSolutions(transitiveClosure(currentNudpeContextSolutionsMaps))
				}
				currentNudpeRecord = er.gNudpeRecord
				currentNudpeContextSolutionsMaps = []contextSolutionsMap{}
			}
		} else {
			if currentNudpeRecord != nil {
				currentNudpeRecord.setContextSolutions(transitiveClosure(currentNudpeContextSolutionsMaps))
				currentNudpeRecord = nil
				currentNudpeContextSolutionsMaps = nil
			}
		}

		er.stopUnfoundedSpeculativeExecutionThreads(executor.nudpeBooleanValueEvaluator)
		er.produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads()

		if er.belongsToNudpe() {
			currentNudpeContextSolutionsMaps = append(currentNudpeContextSolutionsMaps, er.ctxSols)
		}
	}

	if currentNudpeRecord != nil {
		currentNudpeRecord.setContextSolutions(transitiveClosure(currentNudpeContextSolutionsMaps))
	}

	return
}

func (executor *executor) nudpeBooleanValueEvaluator(udpeID int, context *NonTerminal, evaluationsCount int) customBool {
	record, err := executor.resultingExecutionTable.recordByID(udpeID)

	if err != nil {
		panic(fmt.Sprintf(`cannot retrieve udpe record from resulting execution table: %v`, err))
	}

	if !record.belongsToNudpe() {
		panic("cannot evaluate nudpe boolean value if first udpe does NOT belong to a nudpe")
	}

	nudpeGlobalRecord := record.gNudpeRecord
	return toCustomBool(nudpeGlobalRecord.hasSolutionsFor(context))
}

func (executor *executor) retrieveResults() (results []Position) {
	var potentiallyDuplicatedResults []Position
	switch executor.mainQueryType {
	case UDPE:
		mainQueryUDPERecord := executor.resultingExecutionTable.mainQueryRecord()
		potentiallyDuplicatedResults = mainQueryUDPERecord.ctxSols.convertToGroupOfSolutionsPositions()
	case NUDPE:
		mainQueryNUDPERecord := nudpeGlobalTable.mainQueryRecord()
		potentiallyDuplicatedResults = mainQueryNUDPERecord.contextSolutions().convertToGroupOfSolutionsPositions()
	default:
		panic("retrieving results error: unknown main query type")
	}

	var uniqueResults = make(map[string]bool, len(potentiallyDuplicatedResults))

	for _, result := range potentiallyDuplicatedResults {
		resultIdentifier := fmt.Sprintf("(%v,%v)", result.Start(), result.End())
		_, isAlreadyConsidered := uniqueResults[resultIdentifier]
		if !isAlreadyConsidered {
			results = append(results, result)
			uniqueResults[resultIdentifier] = true
		}
	}

	return
}

