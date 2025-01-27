package xpath

import (
	"fmt"
	"strings"
)

type rpeBuilderState int

const (
	expectRpeAxis rpeBuilderState = iota
	expectRpeUdpeTest
)

type rpeBuilder struct {
	state            rpeBuilderState
	currentInnerTest *rpeInnerTest
}

func newRpeBuilder() *rpeBuilder {
	return new(rpeBuilder)
}

func (rpeBuilder *rpeBuilder) init() {
	return
}

func (rpeBuilder *rpeBuilder) addUdpeTest(udpeTest udpeTest) (ok bool) {
	if rpeBuilder.state != expectRpeUdpeTest {
		ok = false
		return
	}
	defer func() {
		rpeBuilder.state = expectRpeAxis
	}()

	ok = true
	nextRpeInnerTest := &rpeInnerTest{
		udpeTest:         udpeTest,
		nextRpeInnerTest: rpeBuilder.currentInnerTest,
	}
	rpeBuilder.currentInnerTest = nextRpeInnerTest
	return
}

func (rpeBuilder *rpeBuilder) addAxis(axis axis) (ok bool) {
	if rpeBuilder.state != expectRpeAxis {
		ok = false
		return
	}
	defer func() {
		rpeBuilder.state = expectRpeUdpeTest
	}()

	ok = true
	if rpeBuilder.currentInnerTest == nil {
		return
	}
	if axis == ancestorOrSelf {
		rpeBuilder.currentInnerTest.behindAncestorAxis = true
	}
	return
}

func (rpeBuilder *rpeBuilder) end() (result *rpe) {
	if rpeBuilder.currentInnerTest != nil {
		rpeBuilder.currentInnerTest.isEntry = true
		result = &rpe{
			entryTest: rpeBuilder.currentInnerTest,
		}
	}
	return
}

type rpeInnerTest struct {
	isEntry            bool
	behindAncestorAxis bool
	udpeTest           udpeTest
	nextRpeInnerTest   *rpeInnerTest
}

func (rpeInnerTest *rpeInnerTest) matchWithReductionOf(n interface{}) (predicate *predicate, next, newTest *rpeInnerTest, hasNewTest, ok bool) {
	doesUdpeTestMatches := rpeInnerTest.udpeTest.test(n)
	if rpeInnerTest.isEntry {
		switch {
		case rpeInnerTest.behindAncestorAxis:
			ok = true
			next = rpeInnerTest
			if doesUdpeTestMatches {
				newTest = rpeInnerTest.nextRpeInnerTest
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = rpeInnerTest.nextRpeInnerTest
		default:
			ok = false
		}
	} else {
		switch {
		case rpeInnerTest.behindAncestorAxis && rpeInnerTest.nextRpeInnerTest != nil && rpeInnerTest.nextRpeInnerTest.behindAncestorAxis:
			ok = true
			if doesUdpeTestMatches {
				next = rpeInnerTest.nextRpeInnerTest
			} else {
				next = rpeInnerTest
			}
		case rpeInnerTest.behindAncestorAxis:
			ok = true
			next = rpeInnerTest
			if doesUdpeTestMatches {
				newTest = rpeInnerTest.nextRpeInnerTest
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = rpeInnerTest.nextRpeInnerTest
		default:
			ok = false
		}
	}

	if ok {
		predicate = rpeInnerTest.udpeTest.predicate()
	}
	return
}

func (rpeInnerTest *rpeInnerTest) entry() bool {
	return rpeInnerTest.isEntry
}

func (rpeInnerTest *rpeInnerTest) String() (result string) {
	if rpeInnerTest.behindAncestorAxis {
		result = `\\`
	} else if !rpeInnerTest.isEntry {
		result = `\`
	}
	result += fmt.Sprintf("%v", rpeInnerTest.udpeTest)
	return
}

//concrete rpe path pattern implementation
type rpePathPathPattern struct {
	currentTest *rpeInnerTest
}

func (rpePP *rpePathPathPattern) isEmpty() bool {
	return rpePP.currentTest == nil
}

func (rpePP *rpePathPathPattern) matchWithReductionOf(n interface{}, doUpdate bool) (predicate *predicate, newPathPattern pathPattern, ok bool) {
	if rpePP.isEmpty() {
		panic(`rpe path pattern error: trying a match for an empty path pattern`)
	}

	predicate, next, newTest, hasNewTest, ok := rpePP.currentTest.matchWithReductionOf(n)
	if !ok {
		return
	}

	if !doUpdate {
		return
	}

	rpePP.currentTest = next
	if hasNewTest {
		newPathPattern = &rpePathPathPattern{newTest}
	}
	return
}

func (rpePP *rpePathPathPattern) String() (result string) {
	result = rpeStringifyUtil(rpePP.currentTest, rpePathPatternMode)

	if !strings.HasPrefix(result, `\\`) {
		result = strings.TrimPrefix(result, `\`)
	}
	return
}

//concrete rpe implementation
type rpe struct {
	entryTest *rpeInnerTest
}

func (rpe *rpe) entryPoint() pathPattern {
	return &rpePathPathPattern{rpe.entryTest}
}

func (rpe *rpe) String() string {
	return rpeStringifyUtil(rpe.entryTest, rpeMode)
}

//utils
type rpeStringificationMode int

const (
	rpeMode rpeStringificationMode = iota
	rpePathPatternMode
)

func rpeStringifyUtil(currentTest *rpeInnerTest, mode rpeStringificationMode) (result string) {
	for ct := currentTest; ct != nil; ct = ct.nextRpeInnerTest {
		result += ct.String()
	}
	if mode == rpePathPatternMode {
		result += `\ε`
	}
	return
}
