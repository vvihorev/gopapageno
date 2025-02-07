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
var DEBUG = false

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
	resultingExecutionTable executionTable

	source []byte
	runner *gopapageno.Runner
}

// ExecutorCommand represents a command that can be made by a
// client to execute a XPath query
type ExecutorCommand struct {
	xpathQuery      string
	source          []byte
	numberOfThreads int
}

// Execute specify the XPath query to be executed
func Execute(xpathQuery string) *ExecutorCommand {
	DEBUG = false
	logger = newNopLogger()
	return &ExecutorCommand{
		xpathQuery: xpathQuery,
	}
}

func (executorCommand *ExecutorCommand) WithNumberOfThreads(numberOfThreads int) *ExecutorCommand {
	executorCommand.numberOfThreads = numberOfThreads
	return executorCommand
}

func (executorCommand *ExecutorCommand) InVerboseMode() *ExecutorCommand {
	DEBUG = true
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
	if DEBUG {
		logger.Printf("Executing Query: %v", executorCommand.xpathQuery)
	}

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

func (executor *executor) parseQuery(xpathQuery string) {
	executor.mainQueryType = parse(xpathQuery)
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

	for _, er := range executor.resultingExecutionTable.records {
		if er.gNudpeRecord != nil {
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

		// stop unfounded speculative execution threads
		for i := 0; i < er.threads.size; i++ {
			if areSpeculationsFounded := er.threads.array[i].checkAndUpdateSpeculations(executor.nudpeBooleanValueEvaluator); !areSpeculationsFounded {
				er.removeExecutionThread(er.threads.array[i], true)
			}
		}

		// produce context solutions out of completed non speculative execution threads
		for i := 0; i < er.threads.size; i++ {
			if er.threads.array[i].pp.isEmpty() && !er.threads.array[i].isSpeculative() {
				er.ctxSols.addContextSolution(er.threads.array[i].ctx.Position().Start(), er.threads.array[i].ctx.Position().End(), er.threads.array[i].sol)
				er.removeExecutionThread(er.threads.array[i], false)
			}
		}

		if er.gNudpeRecord != nil {
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

	if record.gNudpeRecord == nil {
		panic("cannot evaluate nudpe boolean value if first udpe does NOT belong to a nudpe")
	}

	nudpeGlobalRecord := record.gNudpeRecord
	return toCustomBool(nudpeGlobalRecord.hasSolutionsFor(context))
}

func (executor *executor) retrieveResults() (results []Position) {
	if DEBUG {
		for _, er := range executor.resultingExecutionTable.records {
			logger.Printf("results record: %v", er.String())
			for i := 0; i < er.threads.size; i++ {
				logger.Printf("  resulting thread: %v", er.threads.array[i].String())
			}
		}
	}

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
