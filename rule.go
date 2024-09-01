package gopapageno

type RuleFlags uint8

const (
	RuleSimple RuleFlags = 1 << iota
	RuleAppend
	RuleCombine
	RuleCyclic
	RulePrefix
)

func (t RuleFlags) Set(flag RuleFlags) RuleFlags {
	return t | flag
}

func (t RuleFlags) Clear(flag RuleFlags) RuleFlags {
	return t &^ flag
}

func (t RuleFlags) Toggle(flag RuleFlags) RuleFlags {
	return t ^ flag
}

func (t RuleFlags) Has(flag RuleFlags) bool {
	return t&flag != 0
}

func (t RuleFlags) String() string {
	switch t {
	case RuleSimple:
		return "RuleSimple"
	case RuleCombine:
		return "RuleCombine"
	case RulePrefix:
		return "RulePrefix"
	case RuleCyclic:
		return "RuleCyclic"
	default:
		return "Unknown"
	}
}

type Rule struct {
	Lhs  TokenType
	Rhs  []TokenType
	Type RuleFlags
}
