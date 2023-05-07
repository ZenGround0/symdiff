package symdiff

import (
	"fmt"
)

// Invariant: expression is checked as internally valid
func Differentiate(v Symbol, exp PolyExp) (*PolyExp, error) {
	if exp.s != nil {
		sDiff, err := DifferentiateSum(v, *exp.s)
		if err != nil {
			return nil, err
		}
		return &PolyExp {
			s: sDiff,
		}, nil
	}

	mDiff, err := DifferentiateMonomial(v, *exp.m)
	if err != nil {
		return nil, err
	}
	return &PolyExp {
		m: mDiff,
	}, nil
}

func DifferentiateMonomial(v Symbol, mon MonomialExp) (*MonomialExp, error) {
	if v != mon.x {
		return nil, fmt.Errorf("Cannot take deriviative d/d%s of polynomial function of different bound variable %s", v, mon.x)
	}
	if mon.n == 0 { // technically this is unnecessary but use standard form n==0 for zero polynomial
		return &MonomialExp{
			a: 0,
			x: mon.x,
			n: 0,
		}, nil
	}
	return &MonomialExp{
		a: mon.a * mon.n,
		x: mon.x,
		n: mon.n -1,
	}, nil
}

func DifferentiateSum(v Symbol, sum SumExp) (*SumExp, error) {
	ret := SumExp{ps: make([]PolyExp, len(sum.ps))}
	for i := range sum.ps {
		diff, err := Differentiate(v, sum.ps[i])
		if err != nil {
			return nil, err
		}
		ret.ps[i] = *diff
	}
	return &ret, nil
}
