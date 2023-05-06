package symdiff

// Invariant: exp is checked as internally valid
func Differentiate(exp PolyExp) PolyExp {
	if exp.s != nil {
		return PolyExp {
			s: DifferentiateSum(*exp.s),
		}
	}
	return PolyExp {
		m: DifferentiateMonomial(*exp.m),
	}
}

func DifferentiateMonomial(mon MonomialExp) *MonomialExp {
	if mon.n == 0 { // technically this is unnecessary but use standard form n==0 for zero polynomial
		return &MonomialExp{
			a: 0,
			x: mon.x,
			n: 0,
		}
	}
	return &MonomialExp{
		a: mon.a * mon.n,
		x: mon.x,
		n: mon.n -1,
	}
}

func DifferentiateSum(sum SumExp) *SumExp {
	ret := SumExp{ps: make([]PolyExp, len(sum.ps))}
	for i := range sum.ps {
		ret.ps[i] = Differentiate(sum.ps[i])
	}
	return &ret
}
