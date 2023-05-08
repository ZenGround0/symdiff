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
	if exp.c != nil {
		return &PolyExp {
			c: &ConstantExp {
				c: 0,
			},
		}, nil
	}
	if exp.p != nil {
		productDiff, err := DifferentiateProduct(v, *exp.p)
		if err != nil {
			return nil, err
		}
		return &PolyExp {
			p: productDiff,
		}, nil

	}
	mDiff, err := DifferentiateMonomial(v, *exp.m)
	if err != nil {
		return nil, err
	}
	return &PolyExp {
		p: mDiff,
	}, nil
}

func DifferentiateMonomial(v Symbol, mon MonomialExp) (*ProductExp, error) {
	if v != mon.x {
		return nil, fmt.Errorf("Cannot take deriviative d/d%s of polynomial function of different bound variable %s", v, mon.x)
	}
	var inner MonomialExp
	var multiplicand int
	if mon.n == 0 { // we could use constant as well but we'll let simplification normalize to keep differentiation simple
		inner = MonomialExp{
			x: mon.x,
			n: 0,
		}
		multiplicand = 0
	} else {
		inner = MonomialExp{
			x: mon.x,
			n: mon.n - 1,
		}
		multiplicand = mon.n
	}
	
	return &ProductExp{
		l: &ConstantExp{
			c: multiplicand,
		},
		r: &PolyExp{
			m: &inner,
		},
	}, nil
}

func DifferentiateProduct(v Symbol, prod ProductExp) (*ProductExp, error) {
	diff, err := Differentiate(v, *prod.r)
	if err != nil {
		return nil, err
	}
	return &ProductExp {
		l: prod.l,
		r: diff,
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

