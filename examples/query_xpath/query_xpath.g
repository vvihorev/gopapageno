%axiom Query

%preamble ParserPreallocMem

%%

// WARNING: semantic action is deleted: Step [Test]
// WARNING: semantic action is deleted: OrExpr [AndExpr]
// WARNING: semantic action is deleted: AndExpr [Factor]
// WARNING: semantic action is deleted: Factor [Path]

Query : Path {
    $$.Value = $1.Value
};

Path : Path CHILD Step {
    $$.Value = appendStep($1.Value.(*peSemValue), $3.Value.(udpeTest), child)
} | Path PARENT Step {
    $$.Value = appendStep($1.Value.(*peSemValue), $3.Value.(udpeTest), parent)
} | Path ANCESTOR Step {
    $$.Value = appendStep($1.Value.(*peSemValue), $3.Value.(udpeTest), ancestorOrSelf)
} | Path DESCENDANT Step {
    $$.Value = appendStep($1.Value.(*peSemValue), $3.Value.(udpeTest), descendantOrSelf)
} | CHILD Step {
    $$.Value = appendStep(nil, $2.Value.(udpeTest), child)
} | PARENT Step {
    $$.Value = appendStep(nil, $2.Value.(udpeTest), parent)
} | ANCESTOR Step {
    $$.Value = appendStep(nil, $2.Value.(udpeTest), ancestorOrSelf)
} | DESCENDANT Step {
    $$.Value = appendStep(nil, $2.Value.(udpeTest), descendantOrSelf)
};

Step : Test {
    $$.Value = $1.Value
} | Test LBR OrExpr RBR {
    switch $1.Value.(type) {
    case *elementTest:
        $1.Value.(*elementTest).pred = $3.Value.(*predicate)
    case *textTest:
        $1.Value.(*elementTest).pred = nil
    default:
        panic("unknown test type")
    }
    $$.Value = $1.Value
};

Test : IDENT {
    $$.Value = newElementTest($1.Value.(string), nil, nil)
} | AT IDENT {
    $$.Value = newElementTest("*", &Attribute{Key: $2.Value.(string)}, nil)
} | AT IDENT EQ STRING {
    $$.Value = newElementTest("*", &Attribute{Key: $2.Value.(string), Value: $4.Value.(string)}, nil)
} | TEXT {
    $$.Value = newTextTest("")
} | TEXT EQ STRING {
    $$.Value = newTextTest($3.Value.(string))
};

OrExpr : AndExpr {
    $$.Value = $1.Value
} | AndExpr OR AndExpr {
	predl := $1.Value.(predicate)
	predr := $3.Value.(predicate)

	node := predNode{op: or()}
	predl.root.parent = &node
	predr.root.parent = &node

	for k, v := range predr.undoneAtoms {
		predl.undoneAtoms[k] = v
	}

	$$.Value = predicate{
		root: &node,
		undoneAtoms: predl.undoneAtoms,
	}
};

AndExpr : Factor {
	$$.Value = $1.Value
} | Factor AND Factor {
	predl := $1.Value.(predicate)
	predr := $3.Value.(predicate)

	node := predNode{op: and()}
	predl.root.parent = &node
	predr.root.parent = &node

	for k, v := range predr.undoneAtoms {
		predl.undoneAtoms[k] = v
	}

	$$.Value = predicate{
		root: &node,
		undoneAtoms: predl.undoneAtoms,
	}
};

Factor : Path {
	var pe_id int
	pe := $1.Value.(peSemValue)
	switch pe.builder.(type) {
	case *fpeBuilder:
		pe_id, _ = udpeGlobalTable.addFpe(pe.builder.(*fpeBuilder).end())
	case *rpeBuilder:
		pe_id, _ = udpeGlobalTable.addRpe(pe.builder.(*rpeBuilder).end())
	default:
		panic("something worong")
	}

	node := predNode{op: atom()}
	$$.Value = predicate{root: &node, undoneAtoms: map[int]*predNode{pe_id: &node}}
} | NOT Path {
	var pe_id int
	pe := $2.Value.(peSemValue)
	switch pe.builder.(type) {
	case *fpeBuilder:
		pe_id, _ = udpeGlobalTable.addFpe(pe.builder.(*fpeBuilder).end())
	case *rpeBuilder:
		pe_id, _ = udpeGlobalTable.addRpe(pe.builder.(*rpeBuilder).end())
	default:
		panic("something worong")
	}

	node := predNode{op: atom()}
	$$.Value = predicate{root: &node, undoneAtoms: map[int]*predNode{pe_id: &node}}
} | LPAR OrExpr RPAR {
	$$.Value = $2.Value
} | NOT LPAR OrExpr RPAR {
	pred := $3.Value.(predicate)

	node := predNode{op: not()}
	pred.root.parent = &node
	pred.root = &node

	$$.Value = pred
};

%%

type peSemValue struct {
	pe_id   int
	builder builder
}

func appendStep(pe *peSemValue, step udpeTest, axis axis) *peSemValue {
        if pe == nil {
		pe = &peSemValue{}
		if axis == child || axis == descendantOrSelf {
			pe.builder = &fpeBuilder{}
		} else {
			pe.builder = &rpeBuilder{}
		}

		pe.builder.addAxis(axis)
		pe.builder.addUdpeTest(step)
		return pe
        }

	switch pe.builder.(type) {
	case *fpeBuilder:
		if axis == parent || axis == ancestorOrSelf {
			pe.pe_id, _ = udpeGlobalTable.addFpe(pe.builder.(*fpeBuilder).end())
			pe.builder = &rpeBuilder{}
		} else {
			pe.builder = pe.builder.(*fpeBuilder)
		}
	case *rpeBuilder:
		if axis == child || axis == descendantOrSelf {
			pe.pe_id, _ = udpeGlobalTable.addRpe(pe.builder.(*rpeBuilder).end())
			pe.builder = &fpeBuilder{}
		} else {
			pe.builder = pe.builder.(*rpeBuilder)
		}
	default:
		panic("something worong")
	}

	pe.builder.addAxis(axis)
	pe.builder.addUdpeTest(step)
	return pe
}

// var parserElementsPools []*gopapageno.Pool[xpath.Element]

// ParserPreallocMem initializes all the memory pools required by the semantic function of the parser.
func ParserPreallocMem(inputSize int, numThreads int) {
    // poolSizePerThread := 10000

    // parserElementsPools = make([]*gopapageno.Pool[xpath.Element], numThreads)
    // for i := 0; i < numThreads; i++ {
    //     parserElementsPools[i] = gopapageno.NewPool[xpath.Element](poolSizePerThread)
    // }
}
