package generator

import (
	"strings"
)

func (p *grammarDescription) deleteCopyRules(rulesDict *rulesDictionary) {
	copySets := make(map[string]*set[string], p.nonterminals.Len())
	for _, nonterminal := range p.nonterminals.Iter {
		copySets[nonterminal] = newSet[string]()
	}

	rhsDict := make(map[string][][]string)

	for _, rule := range p.rules {
		if rule.LHS == p.axiom {
			continue
		}
		// If the ruleDescription produces a single nonterminal token,
		// add it to the copy rules set and remove it from the rules' dictionary.
		// Otherwise, add the ruleDescription to rhsRule.
		if len(rule.RHS) == 1 && p.nonterminals.Contains(rule.RHS[0]) {
			copySets[rule.LHS].Add(rule.RHS[0])
			// TODO(vvihorev): this is where semantic actions of renaming rules are lost
			rulesDict.Remove(rule.RHS)
		} else {
			if _, ok := rhsDict[rule.LHS]; ok {
				rhsDict[rule.LHS] = append(rhsDict[rule.LHS], rule.RHS)
			} else {
				rhsDict[rule.LHS] = [][]string{rule.RHS}
			}
		}
	}

	for hasChanged := true; hasChanged; {
		hasChanged = false
		for _, nonterminal := range p.nonterminals.Iter {
			lenCopySet := copySets[nonterminal].Len()

			iterCopy := copySets[nonterminal].Copy()
			for _, copyRhs := range iterCopy.Iter {
				for _, curNonterm := range copySets[copyRhs].Iter {
					copySets[nonterminal].Add(curNonterm)
				}
			}

			if lenCopySet < copySets[nonterminal].Len() {
				hasChanged = true
			}
		}
	}

	for _, nonterminal := range p.nonterminals.Iter {
		for _, copiedNonterm := range copySets[nonterminal].Iter {
			for _, rhs := range rhsDict[copiedNonterm] {
				// There's no need to specify semantic actions
				// because they are already linked to the proper rhs
				rulesDict.Add(&ruleDescription{
					LHS:    nonterminal,
					RHS:    rhs,
					Action: "",
				})
			}
		}
	}
}

func (p *grammarDescription) deleteRepeatedRHS() {
	newRules := make([]ruleDescription, 0)

	// Create a rules dictionary and add every parsed ruleDescription to it.
	dictRules := newRulesDictionary(len(p.rules))
	for _, rule := range p.rules {
		dictRules.Add(&rule)
	}

	// Create a dictionary that will contain the newly added rules.
	newRulesDict := newRulesDictionary(dictRules.Len())

	// TODO: This always causes problems...
	// p.extractTerminalRules(dictRules, newRulesDict)

	p.deleteCopyRules(dictRules)

	V := dictRules.LHSSets()

	dictRulesForIteration := newRulesDictionary(0)
	loop := true
	for loop {
		// Substitutes RHS keys with newly formatted ones.
		for i, _ := range dictRules.KeysRHS {
			rhs := dictRules.KeysRHS[i]
			lhs := dictRules.ValuesLHS[i]
			action := dictRules.SemActions[i]
			flags := dictRules.Flags[i]
			prefixes := dictRules.Prefixes[i]

			newRuleRHS := p.replaceTokenNames(rhs, V)

			for _, lhs := range lhs.Iter {
				for _, rhs := range newRuleRHS {
					dictRulesForIteration.Add(&ruleDescription{
						LHS:      lhs,
						RHS:      rhs,
						Action:   *action,
						Flags:    flags,
						Prefixes: prefixes,
					})
				}
			}
		}

		valueLHSSets := dictRulesForIteration.LHSSets()
		addedNonterminals := make([]*set[string], 0)

		for _, curNontermSet := range valueLHSSets {
			contained := false
			for _, otherNonTermSet := range V {
				if curNontermSet.Equals(otherNonTermSet) {
					contained = true
					break
				}
			}

			if !contained {
				addedNonterminals = append(addedNonterminals, curNontermSet)
				V = append(V, curNontermSet)
			}
		}

		for i, _ := range dictRulesForIteration.KeysRHS {
			keyRHS := dictRulesForIteration.KeysRHS[i]
			valueLHS := dictRulesForIteration.ValuesLHS[i]
			action := dictRulesForIteration.SemActions[i]
			flags := dictRulesForIteration.Flags[i]
			prefixes := dictRulesForIteration.Prefixes[i]

			for _, curLHS := range valueLHS.Iter {
				newRulesDict.Add(&ruleDescription{
					LHS:      curLHS,
					RHS:      keyRHS,
					Action:   *action,
					Flags:    flags,
					Prefixes: prefixes,
				})
			}
		}

		if len(addedNonterminals) == 0 {
			loop = false
		}
	}

	// TODO: remove unused nonterminals (see cpapageno)

	axiomSet := newSet[string]()
	axiomSet.Add(p.axiom)

	axiomSemAction := "{\n\t$$.Value = $1.Value\n}"

	V = append(V, axiomSet)

	for _, nontermSet := range V {
		if nontermSet.Contains(p.axiom) {
			newRulesDict.Add(&ruleDescription{
				LHS:    p.axiom,
				RHS:    []string{strings.Join(nontermSet.Slice(), "_")},
				Action: axiomSemAction,
			})
		}
	}

	// Create the rules from rulesDictionary
	for i, _ := range newRulesDict.KeysRHS {
		keyRHS := newRulesDict.KeysRHS[i]
		valueLHS := newRulesDict.ValuesLHS[i]
		semAction := newRulesDict.SemActions[i]
		flags := newRulesDict.Flags[i]
		prefixes := newRulesDict.Prefixes[i]

		newPrefixes := make([][]string, 0)
		for _, prefix := range prefixes {
			replacedPrefix := p.replaceTokenNames(prefix, newRulesDict.LHSSets())
			newPrefixes = append(newPrefixes, replacedPrefix...)
		}
		prefixes = newPrefixes

		newRules = append(newRules, ruleDescription{
			LHS:      strings.Join(valueLHS.Slice(), "_"),
			RHS:      keyRHS,
			Action:   *semAction,
			Flags:    flags,
			Prefixes: prefixes,
		})
	}

	p.rules = newRules

	p.inferTokens()
}

func (p *grammarDescription) extractTerminalRules(dictRules *rulesDictionary, newRulesDict *rulesDictionary) {
	// Range over the current rules, check if the RHS contains any nonterminal
	// If it doesn't (i.e. it is a *terminal ruleDescription*), add it to the new rules dictionary and remove it from the old dict.
	dictCopy := dictRules.Copy()
	for i, _ := range dictCopy.KeysRHS {
		keyRHS := dictCopy.KeysRHS[i]
		valueLHS := dictCopy.ValuesLHS[i]
		semAction := dictCopy.SemActions[i]
		isPrefix := dictCopy.Flags[i]

		isTerminalRule := true
		for _, token := range keyRHS {
			if p.nonterminals.Contains(token) {
				isTerminalRule = false
				break
			}
		}

		if isTerminalRule {
			for _, curLHS := range valueLHS.Iter {
				newRulesDict.Add(&ruleDescription{
					LHS:    curLHS,
					RHS:    keyRHS,
					Action: *semAction,
					Flags:  isPrefix,
				})
			}

			dictRules.Remove(keyRHS)
		}
	}
}

func (p *grammarDescription) replaceTokenNames(keyRHS []string, newNonterminals []*set[string]) [][]string {
	newTokenNames := make([][]string, 0)

	var rec func(tokens []string, newTokens []string)
	rec = func(tokens []string, newTokens []string) {
		if len(tokens) == 0 {
			newTokenNames = append(newTokenNames, newTokens)
			return
		}

		token := tokens[0]
		if p.nonterminals.Contains(token) {
			for _, nonTermSuperSet := range newNonterminals {
				if nonTermSuperSet.Contains(token) {
					newTokens = append(newTokens, strings.Join(nonTermSuperSet.Slice(), "_"))
					rec(tokens[1:], newTokens)

					newTokensCopy := make([]string, len(newTokens)-1)
					copy(newTokensCopy, newTokens)
					newTokens = newTokensCopy
				} else {

				}
			}
		} else {
			newTokens = append(newTokens, token)
			rec(tokens[1:], newTokens)

			newTokensCopy := make([]string, len(newTokens)-1)
			copy(newTokensCopy, newTokens)
			newTokens = newTokensCopy
		}
	}

	newTokens := make([]string, 0)
	rec(keyRHS, newTokens)

	return newTokenNames
}

func addNewRules(dict *rulesDictionary, index int, p *grammarDescription, newNonterminals []*set[string]) {

}

func (p *grammarDescription) sortRulesByRHS() {
	sortedRules := make([]ruleDescription, 0, len(p.rules))

	for _, curRule := range p.rules {
		insertPosition := -1
		for i, curSortedRule := range sortedRules {
			if p.rhsLessThan(curRule.RHS, curSortedRule.RHS) {
				insertPosition = i
				break
			}
		}
		if insertPosition == -1 {
			sortedRules = append(sortedRules, curRule)
		} else {
			sortedRules = append(sortedRules, ruleDescription{})
			copy(sortedRules[insertPosition+1:], sortedRules[insertPosition:])
			sortedRules[insertPosition] = curRule
		}
	}

	p.rules = sortedRules
}

func (p *grammarDescription) rhsLessThan(rhs1 []string, rhs2 []string) bool {
	minLen := len(rhs1)
	if len(rhs2) < minLen {
		minLen = len(rhs2)
	}

	for i := 0; i < minLen; i++ {
		//If the first is in nonterminals and the second is in terminals,
		//the first token is certainly less than the second
		if p.nonterminals.Contains(rhs1[i]) && p.terminals.Contains(rhs2[i]) {
			return true
		}
		//If the first is in terminals and the second is in nonterminals,
		//the first token is certainly greater than the second
		if p.terminals.Contains(rhs1[i]) && p.nonterminals.Contains(rhs2[i]) {
			return false
		}

		if rhs1[i] < rhs2[i] {
			return true
		}
		if rhs1[i] > rhs2[i] {
			return false
		}
	}

	if len(rhs1) < len(rhs2) {
		return true
	}
	return false
}
