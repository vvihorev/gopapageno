%axiom Query

%preamble ParserPreallocMem

%%

Query : Path {
    $$.Value = $1.Value
};

Path : Path CHILD Step {
    $$.Value = appendStep($1.Value.(string), $3.Value.(udpeTest), child)
} | Path PARENT Step {
    $$.Value = appendStep($1.Value.(string), $3.Value.(udpeTest), parent)
} | Path ANCESTOR Step {
    $$.Value = appendStep($1.Value.(string), $3.Value.(udpeTest), ancestorOrSelf)
} | Path DESCENDANT Step {
    $$.Value = appendStep($1.Value.(string), $3.Value.(udpeTest), descendantOrSelf)
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
        $1.Value.(*elementTest).pred = $3.Value
    case *textTest:
        $1.Value.(*textTest).pred = $3.Value
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
    $$.Value = predicate{
        expressionVector: []operator{or(), atom(), atom()},
        atomsLookup: map[atomID]int{
            atomID($1.Value): 0,
            atomID($3.Value): 1,
        },
    }
};

AndExpr : Factor {
    $$.Value = $1.Value
} | Factor AND Factor {
    $$.Value = predicate{
        expressionVector: []operator{and(), atom(), atom()},
        atomsLookup: map[atomID]int{
            atomID($1.Value): 0,
            atomID($3.Value): 1,
        },
    }
};

Factor : Path {
    $$.Value = $1.Value
} | NOT Path {
    $$.Value = predicate{
        expressionVector: []operator{not(), atom()},
        atomsLookup: map[atomID]int{
            atomID($2.Value): 0,
        },
    }
} | LPAR OrExpr RPAR {
    $$.Value = $1.Value
} | NOT LPAR OrExpr RPAR {
    $$.Value = predicate{
        expressionVector: []operator{not(), atom()},
        atomsLookup: map[atomID]int{
            atomID($3.Value): 0,
        },
    }
};

%%

func appendStep(curBuilder any, step udpeTest, axis axis) (builder any) {
	switch curBuilder.(type) {
	case *fpeBuilder:
		if axis == parent || axis == ancestorOrSelf {
			udpeGlobalTable.addFpe(curBuilder.(*fpeBuilder).end())
			builder := &rpeBuilder{}
			builder.addAxis(axis)
			builder.addUdpeTest(step)
		} else {
			builder := curBuilder.(*fpeBuilder)
			builder.addAxis(axis)
			builder.addUdpeTest(step)
		}
	case *rpeBuilder:
		if axis == child || axis == descendantOrSelf {
			udpeGlobalTable.addRpe(curBuilder.(*rpeBuilder).end())
			builder := &fpeBuilder{}
			builder.addAxis(axis)
			builder.addUdpeTest(step)
		} else {
			builder := curBuilder.(*rpeBuilder)
			builder.addAxis(axis)
			builder.addUdpeTest(step)
		}
	default:
		if axis == child || axis == descendantOrSelf {
			builder := &fpeBuilder{}
			builder.addAxis(axis)
			builder.addUdpeTest(step)
		} else {
			builder := &rpeBuilder{}
			builder.addAxis(axis)
			builder.addUdpeTest(step)
		}
	}
	return
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
