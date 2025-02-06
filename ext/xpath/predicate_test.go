package xpath

import (
	"reflect"
	"testing"
)

func TestAtomOperator(t *testing.T) {
	atom := atom()

	var tests = []struct {
		input customBool
		want  customBool
	}{
		{True, True},
		{False, False},
		{Undefined, Undefined},
	}

	for _, test := range tests {
		if got := atom.evaluate(test.input); got != test.want {
			t.Errorf(`atom(%v)=%v | want %v`, test.input, got, test.want)
		}
	}
}

func TestNotOperator(t *testing.T) {
	not := not()

	var tests = []struct {
		input customBool
		want  customBool
	}{
		{True, False},
		{False, True},
		{Undefined, Undefined},
	}

	for _, test := range tests {
		if got := not.evaluate(test.input); got != test.want {
			t.Errorf(`not(%v)=%v | want %v`, test.input, got, test.want)
		}
	}
}

func TestAndOperator(t *testing.T) {
	t.Run(`and(False)=False`, func(t *testing.T) {
		and := and()
		if got := and.evaluate(False); got != False {
			t.Errorf(`and(False)=%v | want False`, got)
		}
	})

	t.Run(`and(True)=Undefined -> and(False)=False`, func(t *testing.T) {
		and := and()
		if got := and.evaluate(True); got != Undefined {
			t.Errorf(`and(True)=%v | want Undefined`, got)
		}

		if got := and.evaluate(False); got != False {
			t.Errorf(`and(False)=%v | want False`, got)
		}
	})

	t.Run(`and(True)=Undefined -> and(True)=True`, func(t *testing.T) {
		and := and()
		if got := and.evaluate(True); got != Undefined {
			t.Errorf(`and(True)=%v | want Undefined`, got)
		}

		if got := and.evaluate(True); got != True {
			t.Errorf(`and(True)=%v | want True`, got)
		}
	})
}

func TestOrOperator(t *testing.T) {
	t.Run(`or(True)=True`, func(t *testing.T) {
		or := or()
		if got := or.evaluate(True); got != True {
			t.Errorf(`or(True)=%v | want True`, got)
		}
	})

	t.Run(`or(False)=Undefined -> or(True)=True`, func(t *testing.T) {
		or := or()
		if got := or.evaluate(False); got != Undefined {
			t.Errorf(`or(False)=%v | want Undefined`, got)
		}

		if got := or.evaluate(True); got != True {
			t.Errorf(`or(True)=%v | want True`, got)
		}
	})

	t.Run(`or(False)=Undefined -> or(False)=False`, func(t *testing.T) {
		or := or()
		if got := or.evaluate(False); got != Undefined {
			t.Errorf(`or(False)=%v | want Undefined`, got)
		}

		if got := or.evaluate(False); got != False {
			t.Errorf(`or(False)=%v | want False`, got)
		}
	})
}

func TestPredicate(t *testing.T) {
	var F, E, H, A *predNode

	var predicateBuilder = func() (p predicate) {
		//p(A,E,F,H) = -F and E and (H or A)
		n4 := predNode{op: and()}

		n3 := predNode{op: not(), parent: &n4}
		F = &predNode{op: atom(), parent: &n3}

		n1 := predNode{op: and(), parent: &n4}
		E = &predNode{op: atom(), parent: &n1}

		n2 := predNode{op: or(), parent: &n1}
		A = &predNode{op: atom(), parent: &n2}
		H = &predNode{op: atom(), parent: &n2}

		p = predicate{
			root: &n4,
			undoneAtoms: []*predNode{F, E, H, A},
		}
		return p
	}
	t.Run(`earlyEvaluation`, func(t *testing.T) {
		t.Run(`p.earlyEvaluate(F, True)=False -> ... -> p.earlyEvaluate(_, _)=False`, func(t *testing.T) {
			p := predicateBuilder()
			var evaluations = []struct {
				atom   *predNode
				value  customBool
				want   customBool
			}{
				{F, True, False},
				{A, True, False},
				{E, False, False},
				{H, False, False},
			}
			for _, evaluation := range evaluations {
				if got := p.earlyEvaluate(evaluation.atom, evaluation.value); got != evaluation.want {
					t.Errorf(`p.earlyEvaluate(%v, %v)=%v | want %v`, evaluation.atom, evaluation.value, got, evaluation.want)
				}
			}
		})

		t.Run(`p.earlyEvaluation(E, False)=False -> ... -> p.earlyEvaluate(_, _)=False`, func(t *testing.T) {
			p := predicateBuilder()
			var evaluations = []struct {
				atom   *predNode
				value  customBool
				want   customBool
			}{
				{E, False, False},
				{F, False, False},
				{A, False, False},
				{H, True, False},
			}

			for _, evaluation := range evaluations {
				if got := p.earlyEvaluate(evaluation.atom, evaluation.value); got != evaluation.want {
					t.Errorf(`p.earlyEvaluate(%v, %v)=%v | want %v`, evaluation.atom, evaluation.value, got, evaluation.want)
				}
			}
		})

		t.Run(`p.earlyEvaluate(_, _)=Undefined -> ... -> p.earlyEvaluate(F, False)=True`, func(t *testing.T) {
			p := predicateBuilder()
			var evaluations = []struct {
				atom   *predNode
				value  customBool
				want   customBool
			}{
				{H, True, Undefined},
				{A, False, Undefined},
				{E, True, Undefined},
				{F, False, True},
			}
			for _, evaluation := range evaluations {
				if got := p.earlyEvaluate(evaluation.atom, evaluation.value); got != evaluation.want {
					t.Errorf(`p.earlyEvaluate(%v, %v)=%v | want %v`, evaluation.atom, evaluation.value, got, evaluation.want)
				}
			}

		})
	})

	t.Run(`copy`, func(t *testing.T) {
		p := predicateBuilder()
		pc := p.copy()

		if !reflect.DeepEqual(p, *pc) {
			t.Error(`p.copy() not deep equal to p`)
		}
	})

	t.Run(`atomsIDs`, func(t *testing.T) {
		p := predicateBuilder()
		want := []*predNode{F, E, H, A}
		got := p.undoneAtoms

		if lenGot, lenWant := len(got), len(want); lenGot != lenWant {
			t.Errorf(`len(p.atomsIDs())=%d | want %d`, lenGot, lenWant)
		}

		for _, atomID := range want {
			if !contains(got, atomID) {
				t.Errorf(`p.atomsIDs()=%v | want %v`, got, want)
				break
			}
		}
	})
}

//utils
func contains(s []*predNode, e *predNode) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
