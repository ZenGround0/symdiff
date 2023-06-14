package symdiff

import (
	"sort"
)

/*
simplification logic
- de nest all sums into one flat sum expression
- distribute products through all poly, sum, products and constants, only keep around monomials
- add together all monomials of the same term
- normalizze ( ^ x 0) to constant 1
- drop zero constants
*/
func Simplify(poly PolyExp) (*PolyExp, error) {
	// Distribute
	// All products are distributed through to constant or monomial terms
	poly1, err := ApplyProducts(1, poly)
	if err != nil {
		return nil, err
	}

	// Flatten
	// All terms are flattened to one sum
	// Add in zero term if there is only one
	terms := Flatten(*poly1)

	// Fold
	// All terms of the same exponent and symbol are added together
	terms = Fold(terms)
	// Drop
	// Zero constant is removed from top level if there are any other terms
	terms = DropZero(terms)

	// Wrap terms into one flat sum
	// Noop if just one term
	return Join(terms), nil
}

func DropZero(polys []PolyExp) []PolyExp {
	// If there is only one term it can be the zero constant
	if len(polys) < 2 {
		return polys
	}
	nonzero := make([]PolyExp, 0, len(polys))
	for _, poly := range polys {
		if poly.IsConstant() && poly.c.c == 0 {
			continue
		}
		nonzero = append(nonzero, poly)
	}
	return nonzero
}

// Combine monomials with same variable and order adding coefficients
// Skips sums and products with non-monomial right terms.  To do a full
// reduction into component monomials this should be applied after
// ApplyProducts and Flatten.
func Fold(polys []PolyExp) []PolyExp {
	coefficients := make(map[Symbol]map[int]int) // ( * a ( ^ x n ) ) ==> map[x]->map[n]->a
	constantCoeff := 0
	addCoeff := func(n, a int, sym Symbol) {
		if _, ok := coefficients[sym]; !ok {
			coefficients[sym] = make(map[int]int)
		}
		if _, ok := coefficients[sym][n]; !ok {
			coefficients[sym][n] = 0
		}
		coefficients[sym][n] += a
	}

	terms := make([]PolyExp, 0)
	for _, poly := range polys {
		if poly.IsSum() {
			terms = append(terms, poly) // untransformed terms
		} else if poly.IsProduct() {
			if poly.p.r.IsMon() {
				a := poly.p.l.c
				sym := poly.p.r.m.x
				n := poly.p.r.m.n
				addCoeff(n, a, sym)
			} else if poly.p.r.IsConstant() {
				a := poly.p.l.c * poly.p.r.c.c
				constantCoeff += a
			} else {
				terms = append(terms, poly)
			}
		} else if poly.IsMon() {
			addCoeff(poly.m.n, 1, poly.m.x)
		} else { // constant case
			constantCoeff += poly.c.c
		}
	}
	syms := make([]string, 0)
	for sym := range coefficients {
		syms = append(syms, string(sym))
	}
	sort.Strings(syms)

	for _, s := range syms {
		sym := Symbol(s)
		powers := make([]int, 0)
		for power := range coefficients[sym] {
			powers = append(powers, power)
		}
		sort.Ints(powers)
		for _, power := range powers {
			a := coefficients[sym][power]

			if power == 0 {
				constantCoeff += a
				continue
			}

			m := MonomialExp{
				x: sym,
				n: power,
			}
			if a == 1 {
				terms = append(terms, PolyExp{m: &m})
			} else {
				mon := PolyExp{
					p: &ProductExp{
						l: &ConstantExp{
							c: a,
						},
						r: &PolyExp{
							m: &m,
						},
					},
				}
				terms = append(terms, mon)
			}
		}
	}
	terms = append(terms, PolyExp{c: &ConstantExp{c: constantCoeff}})

	return terms
}

func Flatten(poly PolyExp) []PolyExp {
	if poly.IsConstant() || poly.IsMon() {
		return []PolyExp{poly}
	}

	// Flatten does not recurse over products
	// Its intended use is over polynomials that have already distributed all products
	if poly.IsProduct() {
		return []PolyExp{poly}
	}

	// Sum case
	flattened := make([]PolyExp, 0)
	for _, p := range poly.s.ps {
		flattened = append(flattened, Flatten(p)...)
	}
	return flattened
}

func Zero() PolyExp {
	return PolyExp{
		c: &ConstantExp{
			c: 0,
		},
	}
}

func Join(polys []PolyExp) *PolyExp {
	if len(polys) == 1 {
		return &polys[0]
	}
	return &PolyExp{
		s: &SumExp{
			ps: polys,
		},
	}
}

func ApplyProducts(mult int, poly PolyExp) (*PolyExp, error) {
	if mult == 0 {
		return &PolyExp{
			c: &ConstantExp{
				c: 0,
			},
		}, nil
	}
	if poly.IsConstant() {
		poly.c.c *= mult
		return &poly, nil
	}

	if poly.IsMon() {
		return &PolyExp{
			p: &ProductExp{
				l: &ConstantExp{
					c: mult,
				},
				r: &poly,
			},
		}, nil
	}

	if poly.IsSum() {
		sum := poly.s
		ret := PolyExp{
			s: &SumExp{
				ps: make([]PolyExp, len(sum.ps)),
			},
		}
		for i, p := range sum.ps {
			appliedPoly, err := ApplyProducts(mult, p)
			if err != nil {
				return nil, err
			}
			ret.s.ps[i] = *appliedPoly
		}
		return &ret, nil
	}
	// Product case
	return ApplyProducts(poly.p.l.c*mult, *poly.p.r)
}
