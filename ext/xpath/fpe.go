package xpath

import (
	"fmt"
	"strings"
)

type fpeBuilderState int

const (
	expectFpeAxis fpeBuilderState = iota
	expectFpeUdpeTest
)

//concrete implementation of a fpe builder
type fpeBuilder struct {
	state                 fpeBuilderState
	precedentFpeInnerTest *fpeInnerTest
}

func newFpeBuilder() *fpeBuilder {
	return new(fpeBuilder)
}

func (fpeBuilder *fpeBuilder) init() {
	return
}

func (fpeBuilder *fpeBuilder) addUdpeTest(udpeTest udpeTest) (ok bool) {
	if fpeBuilder.state != expectFpeUdpeTest {
		ok = false
		return
	}
	defer func() {
		fpeBuilder.state = expectFpeAxis
	}()

	ok = true
	nextFpeInnerTest := &fpeInnerTest{
		udpeTest:              udpeTest,
		precedingFpeInnerTest: fpeBuilder.precedentFpeInnerTest,
	}

	fpeBuilder.precedentFpeInnerTest = nextFpeInnerTest
	return
}

func (fpeBuilder *fpeBuilder) addAxis(axis axis) (ok bool) {
	if fpeBuilder.state != expectFpeAxis {
		ok = false
		return
	}
	defer func() {
		fpeBuilder.state = expectFpeUdpeTest
	}()

	ok = true
	if fpeBuilder.precedentFpeInnerTest == nil {
		return
	}

	if axis == descendantOrSelf {
		fpeBuilder.precedentFpeInnerTest.behindDescendantAxis = true
	}
	return
}

func (fpeBuilder *fpeBuilder) end() (result *fpe) {
	if fpeBuilder.precedentFpeInnerTest != nil {
		fpeBuilder.precedentFpeInnerTest.isEntry = true
		result = &fpe{
			entryTest: fpeBuilder.precedentFpeInnerTest,
		}
	}
	return
}

type fpeInnerTest struct {
	udpeTest              udpeTest
	precedingFpeInnerTest *fpeInnerTest
	isEntry               bool
	behindDescendantAxis  bool
}

func (fpeInnerTest *fpeInnerTest) matchWithReductionOf(n interface{}) (predicate *predicate, next, newTest *fpeInnerTest, hasNewTest, ok bool) {
	doesUdpeTestMatches := fpeInnerTest.udpeTest.test(n)
	if fpeInnerTest.isEntry {
		switch {
		case fpeInnerTest.behindDescendantAxis:
			ok = true
			next = fpeInnerTest
			if doesUdpeTestMatches {
				newTest = fpeInnerTest.precedingFpeInnerTest
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = fpeInnerTest.precedingFpeInnerTest
		default:
			ok = false
		}
	} else {
		switch {
		case fpeInnerTest.behindDescendantAxis && fpeInnerTest.precedingFpeInnerTest != nil && fpeInnerTest.precedingFpeInnerTest.behindDescendantAxis:
			ok = true
			if doesUdpeTestMatches {
				next = fpeInnerTest.precedingFpeInnerTest
			} else {
				next = fpeInnerTest
			}
		case fpeInnerTest.behindDescendantAxis:
			ok = true
			next = fpeInnerTest
			if doesUdpeTestMatches {
				newTest = fpeInnerTest.precedingFpeInnerTest
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = fpeInnerTest.precedingFpeInnerTest
		default:
			ok = false
		}
	}
	if ok {
		predicate = fpeInnerTest.udpeTest.predicate()
	}
	return
}

func (fpeInnerTest *fpeInnerTest) String() (result string) {
	result = fmt.Sprintf("%v", fpeInnerTest.udpeTest)
	if fpeInnerTest.behindDescendantAxis {
		result += "//"
	} else if !fpeInnerTest.isEntry {
		result += "/"
	}
	return
}

//concrete fpe path pattern implementation
type fpePathPattern struct {
	currentTest *fpeInnerTest
}

func (fpePP *fpePathPattern) isEmpty() bool {
	return fpePP.currentTest == nil
}

func (fpePP *fpePathPattern) matchWithReductionOf(n interface{}, doUpdate bool) (predicate *predicate, newPathPattern pathPattern, ok bool) {
	if fpePP.isEmpty() {
		panic(`fpe path pattern error: trying a match for an empty path pattern`)
	}

	predicate, next, newTest, hasNewTest, ok := fpePP.currentTest.matchWithReductionOf(n)
	if !ok {
		return
	}

	if !doUpdate {
		return
	}

	fpePP.currentTest = next
	if hasNewTest {
		newPathPattern = &fpePathPattern{newTest}
	}
	return
}

func (fpePP *fpePathPattern) String() (result string) {
	result = fpeStringifyUtil(fpePP.currentTest, fpePathPatternMode)

	if !strings.HasSuffix(result, "//") {
		result = strings.TrimSuffix(result, "/")
	}
	return
}

//concrete fpe implementation
type fpe struct {
	udpe
	entryTest *fpeInnerTest
}

func (fpe *fpe) entryPoint() pathPattern {
	return &fpePathPattern{fpe.entryTest}
}

func (fpe *fpe) String() string {
	return fpeStringifyUtil(fpe.entryTest, fpeMode)
}

//Utils
type fpeStringificationMode int

const (
	fpeMode fpeStringificationMode = iota
	fpePathPatternMode
)

func fpeStringifyUtil(currentTest *fpeInnerTest, mode fpeStringificationMode) (result string) {
	if currentTest != nil {
		result = fpeStringifyUtil(currentTest.precedingFpeInnerTest, mode) + fmt.Sprintf("%v", currentTest)
	} else {
		if mode == fpePathPatternMode {
			result = "ε/"
		}
	}
	return
}
