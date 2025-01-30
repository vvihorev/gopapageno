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

// concrete implementation of a fpe builder
type fpeBuilder struct {
	state fpeBuilderState
	head  *fpeInnerTest
}

func (fpeBuilder *fpeBuilder) addUdpeTest(udpeTest udpeTest) {
	if fpeBuilder.state != expectFpeUdpeTest {
		panic("expected to find a UDPE Test in the query")
	}
	fpeBuilder.state = expectFpeAxis

	test := &fpeInnerTest{
		udpeTest: udpeTest,
		next:     fpeBuilder.head,
	}
	fpeBuilder.head = test
}

func (fpeBuilder *fpeBuilder) addAxis(axis axis) {
	if fpeBuilder.state != expectFpeAxis {
		panic("expected to find an Axis in the query")
	}
	fpeBuilder.state = expectFpeUdpeTest

	if fpeBuilder.head != nil && axis == descendantOrSelf {
		fpeBuilder.head.behindDescendantAxis = true
	}
}

func (fpeBuilder *fpeBuilder) end() *fpe {
	if fpeBuilder.head == nil {
		panic("attempted to build an empty FPE (no UDPE Tests created)")
	}
	fpeBuilder.head.isEntry = true
	return &fpe{
		entryTest: fpeBuilder.head,
	}
}

// fpeBuilder returns a linked list of fpeInnerTests.
// e.g. the input query '/html//div/form//button' becomes:
//   HEAD
//   button -->  form  -->  div  -->  html
//   isEntry    bhDscA               bhDscA
type fpeInnerTest struct {
	udpeTest             udpeTest
	next                 *fpeInnerTest
	isEntry              bool
	behindDescendantAxis bool
}

func (fpeInnerTest *fpeInnerTest) matchWithReductionOf(n interface{}) (predicate *predicate, next, newTest *fpeInnerTest, hasNewTest, ok bool) {
	udpeTestMatches := fpeInnerTest.udpeTest.test(n)
	if fpeInnerTest.isEntry {
		switch {
		case fpeInnerTest.behindDescendantAxis:
			ok = true
			next = fpeInnerTest
			if udpeTestMatches {
				newTest = fpeInnerTest.next
				hasNewTest = true
			}
		case udpeTestMatches:
			ok = true
			next = fpeInnerTest.next
		default:
			ok = false
		}
	} else {
		switch {
		case fpeInnerTest.behindDescendantAxis && fpeInnerTest.next != nil && fpeInnerTest.next.behindDescendantAxis:
			ok = true
			if udpeTestMatches {
				next = fpeInnerTest.next
			} else {
				next = fpeInnerTest
			}
		case fpeInnerTest.behindDescendantAxis:
			ok = true
			next = fpeInnerTest
			if udpeTestMatches {
				newTest = fpeInnerTest.next
				hasNewTest = true
			}
		case udpeTestMatches:
			ok = true
			next = fpeInnerTest.next
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

// concrete fpe path pattern implementation
type fpePathPattern struct {
	head *fpeInnerTest
}

func (fpePP *fpePathPattern) isEmpty() bool {
	return fpePP.head == nil
}

func (fpePP *fpePathPattern) matchWithReductionOf(n interface{}, doUpdate bool) (predicate *predicate, newPathPattern pathPattern, ok bool) {
	if fpePP.isEmpty() {
		panic(`fpe path pattern error: trying a match for an empty path pattern`)
	}

	predicate, next, newTest, hasNewTest, ok := fpePP.head.matchWithReductionOf(n)
	if !ok {
		return
	}

	if !doUpdate {
		return
	}

	fpePP.head = next
	if hasNewTest {
		newPathPattern = &fpePathPattern{newTest}
	}
	return
}

func (fpePP *fpePathPattern) String() (result string) {
	result = fpeStringifyUtil(fpePP.head, fpePathPatternMode)

	if !strings.HasSuffix(result, "//") {
		result = strings.TrimSuffix(result, "/")
	}
	return
}

// Forward Path Expression, build a FPE and match nodes of the parse tree against it
//
// The FPE is a path pattern - linked list of path steps.
//
// Each step has an axis: either matching a direct child node, or a
// descendant-or-self node.
//
// Moreover, each step has a test: a tag name, and/or an attribute value,
// and/or a predicate expression to check.
//
// FPE is a UDPE. FPE provides acces to its entry point's path pattern.
// The path pattern is a linked list of tests: predicates which can be evaluated for a given node.
// Tests supported at the moment are text tests and element tests. Element tests include tag identifier,
// attribute, and predicate tests.
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

// Utils
type fpeStringificationMode int

const (
	fpeMode fpeStringificationMode = iota
	fpePathPatternMode
)

func fpeStringifyUtil(currentTest *fpeInnerTest, mode fpeStringificationMode) (result string) {
	if currentTest != nil {
		result = fpeStringifyUtil(currentTest.next, mode) + fmt.Sprintf("%v", currentTest)
	} else {
		if mode == fpePathPatternMode {
			result = "ε/"
		}
	}
	return
}
