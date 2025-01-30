package xpath

import (
	"testing"
)

func TestOneTagFPE(t *testing.T) {
	executor := executor{}
	executor.parseQuery("//div")

	if len(udpeGlobalTable.list) != 1 {
		t.Fatalf("Expected one UDPE in the global table")
	}

	record := udpeGlobalTable.list[0]
	if record.expType != FPE {
		t.Fatalf("expected an FPE record, got: %s", udpeType(record.expType))
	}	

	if record.gNudpeRecord != nil {
		t.Fatalf("expected the expression not to be linked to a NUDPE")
	}

	exp, ok := record.exp.(*fpe)
	if !ok {
		t.Fatalf("expected the expression to be an instance of fpe{}")
	}

	head := exp.entryTest
	if !head.isEntry {
		t.Fatalf("entry point not marked as entry")
	}
	if head.behindDescendantAxis {
		t.Fatalf("entry point expected to be marked as not behind descendant axis")
	}

	tag, ok := head.udpeTest.(*elementTest);
	if !ok {
		t.Fatalf("expected to find an element as the entryPoint")
	}
	if tag.name != "div" {
		t.Fatalf("expected to find a 'div' element as the entryPoint")
	}
	if tag.pred != nil{
		t.Fatalf("expected the element to have no predicates")
	}
	if tag.attr != nil{
		t.Fatalf("expected the element to have no attribute tests")
	}

	if head.next != nil{
		t.Fatalf("expected to have only one step in path pattern")
	}
}

func TestTwoTagFPEWithDescendant(t *testing.T) {
	executor := executor{}
	executor.parseQuery("/table//td")

	if len(udpeGlobalTable.list) != 1 {
		t.Fatalf("Expected one UDPE in the global table")
	}

	record := udpeGlobalTable.list[0]
	if record.expType != FPE {
		t.Fatalf("expected an FPE record, got: %s", udpeType(record.expType))
	}	

	if record.gNudpeRecord != nil {
		t.Fatalf("expected the expression not to be linked to a NUDPE")
	}

	exp, ok := record.exp.(*fpe)
	if !ok {
		t.Fatalf("expected the expression to be an instance of fpe{}")
	}

	head := exp.entryTest
	if !head.isEntry {
		t.Fatalf("entry point not marked as entry")
	}
	if head.behindDescendantAxis {
		t.Fatalf("entry point expected to be marked as not behind descendant axis")
	}
	tag, ok := head.udpeTest.(*elementTest);
	if !ok {
		t.Fatalf("expected to find an element in the path pattern")
	}
	if tag.name != "td" {
		t.Fatalf("expected to find a 'div' element as the entryPoint")
	}

	head = head.next
	if head.isEntry {
		t.Fatalf("left elem marked as entry")
	}
	if !head.behindDescendantAxis {
		t.Fatalf("left elem expected to be marked as behind descendant axis")
	}
	tag, ok = head.udpeTest.(*elementTest);
	if !ok {
		t.Fatalf("expected to find two elements in the path pattern")
	}
	if tag.name != "table" {
		t.Fatalf("expected to find a 'div' element as the entryPoint")
	}
	if head.next != nil{
		t.Fatalf("expected to have only one step in path pattern")
	}
}

func TestOneTagRPE(t *testing.T) {
	executor := executor{}
	executor.parseQuery("\\\\head")

	if len(udpeGlobalTable.list) != 1 {
		t.Fatalf("Expected one UDPE in the global table")
	}

	record := udpeGlobalTable.list[0]
	if record.expType != RPE {
		t.Fatalf("expected an RPE record, got: %s", udpeType(record.expType))
	}	

	if record.gNudpeRecord != nil {
		t.Fatalf("expected the expression not to be linked to a NUDPE")
	}

	exp, ok := record.exp.(*rpe)
	if !ok {
		t.Fatalf("expected the expression to be an instance of rpe{}")
	}

	head := exp.entryTest
	if !head.isEntry {
		t.Fatalf("entry point not marked as entry")
	}
	if head.behindAncestorAxis {
		t.Fatalf("entry point expected to be marked as not behind descendant axis")
	}

	tag, ok := head.udpeTest.(*elementTest);
	if !ok {
		t.Fatalf("expected to find an element as the entryPoint")
	}
	if tag.name != "head" {
		t.Fatalf("expected to find a 'head' element as the entryPoint")
	}
	if tag.pred != nil{
		t.Fatalf("expected the element to have no predicates")
	}
	if tag.attr != nil{
		t.Fatalf("expected the element to have no attribute tests")
	}

	if head.next != nil{
		t.Fatalf("expected to have only one step in path pattern")
	}
}

func TestTwoTagRPEWithAncestor(t *testing.T) {
	executor := executor{}
	executor.parseQuery("\\script\\\\head")

	if len(udpeGlobalTable.list) != 1 {
		t.Fatalf("Expected one UDPE in the global table")
	}

	record := udpeGlobalTable.list[0]
	if record.expType != RPE {
		t.Fatalf("expected an RPE record, got: %s", udpeType(record.expType))
	}	

	if record.gNudpeRecord != nil {
		t.Fatalf("expected the expression not to be linked to a NUDPE")
	}

	exp, ok := record.exp.(*rpe)
	if !ok {
		t.Fatalf("expected the expression to be an instance of rpe{}")
	}

	head := exp.entryTest
	if !head.isEntry {
		t.Fatalf("entry point not marked as entry")
	}
	if head.behindAncestorAxis {
		t.Fatalf("entry point expected to be marked as not behind descendant axis")
	}
	tag, ok := head.udpeTest.(*elementTest);
	if !ok {
		t.Fatalf("expected to find an element in the path pattern")
	}
	if tag.name != "script" {
		t.Fatalf("expected to find a 'script' element as the entryPoint")
	}

	head = head.next
	if head.isEntry {
		t.Fatalf("left elem marked as entry")
	}
	if !head.behindAncestorAxis {
		t.Fatalf("left elem expected to be marked as behind descendant axis")
	}
	tag, ok = head.udpeTest.(*elementTest);
	if !ok {
		t.Fatalf("expected to find two elements in the path pattern")
	}
	if tag.name != "head" {
		t.Fatalf("expected to find a 'head' element as the entryPoint")
	}
	if head.next != nil{
		t.Fatalf("expected to have only one step in path pattern")
	}
}
