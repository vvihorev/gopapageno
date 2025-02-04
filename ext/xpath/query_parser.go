package xpath

func parse(xpathQuery string) mainQueryType {
	return parseDPE(xpathQuery, 0, len(xpathQuery))
}

func parseDPE(xpathQuery string, start, end int) mainQueryType {
	var curFpe *fpeBuilder
	var curRpe *rpeBuilder
	hasFpe := false
	hasRpe := false
	udpeCount := 0
	i := start

	if end == start {
		panic("expected a non-empty xpath query")
	}

	udpeGlobalTable = new(globalUdpeTable)
	nudpeGlobalTable = new(globalNudpeTable)

	peek := func() byte {
		if i+1 >= end {
			return 0
		}
		return xpathQuery[i+1]
	}

	readTag := func() string {
		start := i
		for i < end && xpathQuery[i] != '[' && xpathQuery[i] != '/' && xpathQuery[i] != '\\' {
			i++
		}
		if i == start {
			panic("malformed query: expected to find a tag name in Test")
		}
		return xpathQuery[start:i]
	}

	readAttribute := func() *Attribute {
		i++  // skip the initial '@' symbol
		start := i
		for i < end && xpathQuery[i] != ']' && xpathQuery[i] != '=' {
			i++
		}
		if i == end {
			panic("malformed query: expected to find a closing ']' bracket")
		}
		name := xpathQuery[start:i]
		if xpathQuery[i] == '=' {
			i++
			start = i
			for i < end && xpathQuery[i] != ']' {
				i++
			}
			if i <= start + 1 {
				panic("malformed query: expected a value for the attribute")
			}
			value := xpathQuery[start+1:i-1]  // omitting the quotes around the attribute value
			i++
			return &Attribute{Key: name, Value: value}
		}
		i++
		return &Attribute{Key: name}
	}

	readStep := func() (tag string, attribute *Attribute, pred *predicate) {
		tag = readTag()
		// TODO(vvihorev): support text() builtin and predicates, maybe use gopapageno to generate the query parser?
		// TODO(vvihorev): support NUDPEs embedded in predicates
		// TODO(vvihorev): support predicates
		// p := &predicate{
		// 	expressionVector: []operator{and(), atom(), atom()},
		// 	atomsLookup: map[atomID]int{
		// 		atomID(fpe1ID): 1,
		// 		atomID(fpe2ID): 2,
		// 	},
		// }

		if i < end && xpathQuery[i] == '[' {
			switch peek() {
			case '@':
				i++  // step onto the '@' symbol
				attribute = readAttribute()
			default:
				panic("unexpected expression in predicate")
			}
		} else {
			attribute = nil
		}
		return
	}

	for i < end {
		c := xpathQuery[i]

		switch c {
		case '/':
			hasFpe = true
			if curFpe == nil {
				curFpe = &fpeBuilder{}
			}
			if curRpe != nil {
				udpeGlobalTable.addRpe(curRpe.end())
				udpeCount++
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

			curFpe.addUdpeTest(newElementTest(readStep()))

		case '\\':
			hasRpe = true
			if curRpe == nil {
				curRpe = &rpeBuilder{}
			}
			if curFpe != nil {
				udpeGlobalTable.addFpe(curFpe.end())
				udpeCount++
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

			curRpe.addUdpeTest(newElementTest(readStep()))

		default:
			panic("malformed query: expected an Axis")
		}
	}

	if curFpe != nil {
		udpeGlobalTable.addFpe(curFpe.end())
		udpeCount++
	}
	if curRpe != nil {
		udpeGlobalTable.addRpe(curRpe.end())
		udpeCount++
	}

	if hasFpe && hasRpe {
		globalNudpe := nudpeGlobalTable.addNudpeRecord(udpeCount)
		for i := range udpeGlobalTable.size() {
			udpeGlobalTable.list[i].setNudpeRecord(globalNudpe)
		}
		return NUDPE
	} else {
		return UDPE
	}
}
