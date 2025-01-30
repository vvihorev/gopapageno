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
	head *rpeInnerTest
	nextTestIsBehindAncestorAxis bool
}

func (rpeBuilder *rpeBuilder) addUdpeTest(udpeTest udpeTest) {
	if rpeBuilder.state != expectRpeUdpeTest {
		panic("expected to find a UDPE Test in the query")
	}
	rpeBuilder.state = expectRpeAxis

	test := &rpeInnerTest{
		udpeTest:         udpeTest,
		behindAncestorAxis: rpeBuilder.nextTestIsBehindAncestorAxis,
	}
	rpeBuilder.nextTestIsBehindAncestorAxis = false

	if rpeBuilder.head == nil {
		rpeBuilder.head = test
		return
	} 
	n := rpeBuilder.head
	for n.next != nil {
		n = n.next 
	}
	n.next = test
}

func (rpeBuilder *rpeBuilder) addAxis(axis axis) {
	if rpeBuilder.state != expectRpeAxis {
		panic("expected to find an Axis in the query")
	}
	rpeBuilder.state = expectRpeUdpeTest

	if rpeBuilder.head != nil && axis == ancestorOrSelf {
		rpeBuilder.nextTestIsBehindAncestorAxis = true
	}
}

func (rpeBuilder *rpeBuilder) end() *rpe {
	if rpeBuilder.head == nil {
		panic("attempted to build an empty RPE (no UDPE Tests created)")
	}
	rpeBuilder.head.isEntry = true
	return &rpe{
		entryTest: rpeBuilder.head,
	}
}

type rpeInnerTest struct {
	udpeTest           udpeTest
	next   *rpeInnerTest
	isEntry            bool
	behindAncestorAxis bool
}

func (rpeInnerTest *rpeInnerTest) matchWithReductionOf(n interface{}) (predicate *predicate, next, newTest *rpeInnerTest, hasNewTest, ok bool) {
	doesUdpeTestMatches := rpeInnerTest.udpeTest.test(n)
	if rpeInnerTest.isEntry {
		switch {
		case rpeInnerTest.behindAncestorAxis:
			ok = true
			next = rpeInnerTest
			if doesUdpeTestMatches {
				newTest = rpeInnerTest.next
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = rpeInnerTest.next
		default:
			ok = false
		}
	} else {
		switch {
		case rpeInnerTest.behindAncestorAxis && rpeInnerTest.next != nil && rpeInnerTest.next.behindAncestorAxis:
			ok = true
			if doesUdpeTestMatches {
				next = rpeInnerTest.next
			} else {
				next = rpeInnerTest
			}
		case rpeInnerTest.behindAncestorAxis:
			ok = true
			next = rpeInnerTest
			if doesUdpeTestMatches {
				newTest = rpeInnerTest.next
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = rpeInnerTest.next
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
type rpePathPattern struct {
	head *rpeInnerTest
}

func (rpePP *rpePathPattern) isEmpty() bool {
	return rpePP.head == nil
}

func (rpePP *rpePathPattern) matchWithReductionOf(n interface{}, doUpdate bool) (predicate *predicate, newPathPattern pathPattern, ok bool) {
	if rpePP.isEmpty() {
		panic(`rpe path pattern error: trying a match for an empty path pattern`)
	}

	predicate, next, newTest, hasNewTest, ok := rpePP.head.matchWithReductionOf(n)
	if !ok {
		return
	}

	if !doUpdate {
		return
	}

	rpePP.head = next
	if hasNewTest {
		newPathPattern = &rpePathPattern{newTest}
	}
	return
}

func (rpePP *rpePathPattern) String() (result string) {
	result = rpeStringifyUtil(rpePP.head, rpePathPatternMode)

	if !strings.HasPrefix(result, `\\`) {
		result = strings.TrimPrefix(result, `\`)
	}
	return
}

//concrete rpe implementation
type rpe struct {
	udpe
	entryTest *rpeInnerTest
}

func (rpe *rpe) entryPoint() pathPattern {
	return &rpePathPattern{rpe.entryTest}
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
	for ct := currentTest; ct != nil; ct = ct.next {
		result += ct.String()
	}
	if mode == rpePathPatternMode {
		result += `\ε`
	}
	return
}
