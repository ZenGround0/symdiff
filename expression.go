package symdiff

import (
	"fmt"
	"strconv"
	"unicode"
)

/*
   TODOs
   - [x] Test S-Exp parsing
   - [x] Complete S-Exp -> PolyExp parsing
   - [x] PolyExp -> S-Exp serialization
   - [x] Differentiation function PolyExp -> PolyExp
   - [x] Test out e2e differentiation
      -- good thing to check is normalization of sum of sums. The following should have the same derivative
         -- (+ (+ (' 1 x 1) (' 1 x 2)) (+ (' 1 x 3) ( ' 1 x 4) )) =
         -- (+ (' 1 x 1) (' 1 x 2) (' 1 x 3) ( ' 1 x 4) )
         well actually maybe not, you can keep your associative groupings after differentiation as well
         what we should actually do is make a reduction function which normalizes all poly expressions to
         a sum of monomials
         normalization can then be extended in other improvements
      -- might need to normalize poly exp to be a sum of monomial exps but that will make things harder in the future
   6. CLI functionality for taking derivitive
   7. REPL

   **Extension ideas**

   Presentation ideas

   1. Print as Latex
      - idea:
      (+ (' 7 x 2) (1 x 4) (-1 x -8) ) ==>
      `
        \documentclass[12pt]{article}
        \begin{document}
        \[ 7 x^{2} + x^{4} -x^{-8} \]
        \end{document}
      `
      - prep: download latex and play with it working
      - goal: pipe output to a latex renderer and display polynomial
      - goal: support any new expressions added from below extensions
   2. Colorful parentheses matching DONE
   3. REPL -- different commands
     - `d/dx for differentiate
     - `s for simplify
     - `pp for pretty printing maybe as polynomial expression, colored parens, latex etc
   4. Printing polynomial expressions directly as ( + (' 1 x 5 )  (' 2 x 2)) ==> "x^5 + 2x^2

   Differentiation ideas

   1. Chain rule
      - d/dx (f(u)) = df/dx * du/dx
      - prep: think about how to modify expressions + write out a test case
      - goal: add composite polys to S-exp, differentiate fn, test in repl
         - i.e. ((x + 5)^7)^-1  -->
   2. Product rule
      - d/dx (f * g) = df/dx
      - prep: think about how to modify expressions + write out a test case
      - goal: add products to S-exp, differentiate fn, test in repl
        - i.e. (x + 5) * (x^7)  --> (1)x^7 + (x +5)*7x^6
   3. Rational Functions
      - we should do a specialized version of chain rule for power expressions
      - do this plus product rule then we get rational functions and that's cool
      - good goal to shoot for
      - ( / <poly-expr> <poly-expr>) sugar for this internall use product and power expressions
   4. Infinite sums
      - (+inf n (' 1 x n) ), bind a summation variable n and create a term
      - derivative treams n as a constant
   5. Exponential from taylor series
      - need to deal with 1 / n!
      - renormalize n-1 to n
   6. Multi variate derivatives
      - requires getting polynomial expressions to support multivariate
   7. Direct support for other transcendental functions
      - e(x), ln(x), sin(x), cos(x)
      - probably more fun to implement as infinite series of polynomials

   Polynomial expresion ideas

   1. Multivariate polynomials
      - as many bound variables as we want: x,y,z
      - basic atoms of poly expressions become products of monomials with constant a stripped out:
        term = ( a ( ' x 2 ) ( ' y 7 ) ( ' alpha 4 ) )

   2. Simplification of sums of monomials
     - algorithm for simplifying polynomial expressions.
     - drop 0 polynomial terms entirely
     - combine monomials of the same power
     - turn nested sums into single sums

   3. Simplifications of richer expressions
     - Multiply out products
     - Multiply out power expressions
     - Do polynomial division on rational expressions

   4. Expose simplification rules to the repl as a command


*/

/*

   --updated grammar--

   <poly exp>  ::= <sum exp> | <monomial exp> | <product exp> | <constant exp>
   <sum exp> ::= ( sum <poly exp> ... <poly exp> )
   <monomial exp> ::= ( ^ <symbol> <int> )
   <product exp> ::= ( * <constant> <poly exp> )
   <constant exp> ::= int

   simplification logic
   - de nest all sums into one flat sum expression
   - distribute products through all poly, sum, products and constants, only keep around monomials
   - add together all monomials of the same term
   - normalizze ( ^ x 0) to constant 1
   - drop zero constants


*/

/*

   <symbol>       ::= alphabetical string
   <int>          ::= integer string

Syntactic sugar:
sum :=  +
mon :=  ‘

Example

   `5 x^5 + 6 x ^3 - x + 6 + 3 x^-7`
   (+ (‘ 5 x 5) (‘ 6 x 3 ) (‘ -1 x 1) (‘ 6 x 0) (‘ 3 x -7  ))

   Derivative should be

   `25 x ^4 + 18 x ^ 2 - 1 + 0 - 21 x ^-8`

   (+ (‘ 25 x 4) (‘ 18 x 2) (‘ -1 x 0) (‘ -21 x -8))

*/

// datatypes for defining internal expressions of polynomial functions

// Polynomial variables must be purely alphabetic strings
type Symbol string

func IsSymbol(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

const SumKeyWord = "sum"
const SumSugarKeyWord = "+"
const MonomialKeyWord = "mon"
const MonomialSugarKeyWord = "^"
const ProductKeyWord = "prod"
const ProductSugarKeyWord = "*"
const DeprecatedMonomialSyntax = "'"

// Valid atom strings that are not alphanumeric
var SpecialAtoms map[string]struct{}
var Rainbow []int

func init() {
	SpecialAtoms = make(map[string]struct{})
	SpecialAtoms[SumSugarKeyWord] = struct{}{}
	SpecialAtoms[MonomialSugarKeyWord] = struct{}{}
	SpecialAtoms[ProductSugarKeyWord] = struct{}{}
	SpecialAtoms[DeprecatedMonomialSyntax] = struct{}{}

	Rainbow = make([]int, 6)
	Rainbow[1] = 124
	Rainbow[2] = 202
	Rainbow[3] = 11
	Rainbow[4] = 34
	Rainbow[5] = 51
	Rainbow[0] = 19

}

type ConstantExp struct {
	c int
}

func (c *ConstantExp) ToSExp() SExp {
	a := new(Atom)
	*a = Atom(fmt.Sprintf("%d", c.c))
	return SExp{
		Atom: a,
	}
}

func (c *ConstantExp) Parse(s SExp) error {
	if s.Atom == nil {
		return fmt.Errorf("invalid S expression %s, cannot parse as constant polynomial", s.String())
	}
	a, err := strconv.Atoi(string(*s.Atom))
	if err != nil {
		return fmt.Errorf("%s failed to parse constant %s", err, s.String())
	}
	c.c = a

	return nil
}

type ProductExp struct {
	l *ConstantExp
	r *PolyExp
}

func (p *ProductExp) ToSExp() SExp {
	return SExp{
		List: []SExp{
			NewAtom("*"),
			p.l.ToSExp(),
			p.r.ToSExp(),
		},
	}

}

func (p *ProductExp) match(sexp SExp) bool {
	if sexp.Atom == nil {
		return false
	}
	return *sexp.Atom == Atom(ProductKeyWord) || *sexp.Atom == Atom(ProductSugarKeyWord)
}

func (p *ProductExp) Parse(sexp SExp) error {
	if len(sexp.List) != 3 || !p.match(sexp.List[0]) {
		return fmt.Errorf("invalid SExp, cannot parse as polynomial product %s", sexp.String())
	}
	var constExp ConstantExp
	if err := constExp.Parse(sexp.List[1]); err != nil {
		return fmt.Errorf("%s, failed to parse left multiplicand (%s) as constant polynomial", err, sexp.List[1].String())
	}
	p.l = &constExp
	var poly PolyExp
	if err := poly.Parse(sexp.List[2]); err != nil {
		return fmt.Errorf("%s, failed to parse sub expression %s as polynomials", err, sexp.List[2].String())
	}
	p.r = &poly

	return nil
}

type MonomialExp struct {
	x Symbol
	n int
}

// Getter for all fields constituting monomial term
// Fields are private to restrict setting to parsing
func (m *MonomialExp) Term() (Symbol, int) {
	return m.x, m.n
}

// Match monomial to SExp head Atom
func (m *MonomialExp) match(s SExp) bool {
	if s.Atom == nil {
		return false
	}
	return *s.Atom == Atom(MonomialKeyWord) || *s.Atom == Atom(MonomialSugarKeyWord)
}

func (m *MonomialExp) ToSExp() SExp {
	return SExp{
		List: []SExp{
			NewAtom("^"),
			NewAtom(string(m.x)),
			NewAtom(fmt.Sprintf("%d", m.n)),
		},
	}
}

func (m *MonomialExp) Parse(s SExp) error {
	if len(s.List) != 3 && !m.match(s.List[0]) {
		return fmt.Errorf("invalid SExp, cannot parse as monomial %s", s.String())
	}
	for _, sexp := range s.List[1:] {
		if sexp.Atom == nil {
			return fmt.Errorf("invalid SExp, cannot parse as monomial %s", s.String())
		}
	}
	if !IsSymbol(string(*s.List[1].Atom)) {
		return fmt.Errorf("failed to parse variable, not a valid symbol for monomial %s", s.String())
	}
	m.x = Symbol(*s.List[1].Atom)
	n, err := strconv.Atoi(string(*s.List[2].Atom))
	if err != nil {
		return fmt.Errorf("failed to parse exponent %s for monomial %s", err, s.String())
	}
	m.n = n
	return nil
}

type SumExp struct {
	ps []PolyExp
}

// Getter for all sum terms
// Fields are private to restrict setting to parsing
func (s *SumExp) Term() []PolyExp {
	return s.ps
}

func (sum *SumExp) match(sexp SExp) bool {
	if sexp.Atom == nil {
		return false
	}
	return *sexp.Atom == Atom(SumKeyWord) || *sexp.Atom == Atom(SumSugarKeyWord)
}

func (sum *SumExp) ToSExp() SExp {
	sub := []SExp{NewAtom("+")}
	for _, p := range sum.ps {
		sub = append(sub, p.ToSExp())
	}

	return SExp{
		List: sub,
	}
}

func (sum *SumExp) Parse(sexp SExp) error {
	if len(sexp.List) < 3 || !sum.match(sexp.List[0]) {
		return fmt.Errorf("invalid SExp, cannot parse as polynomial sum %s", sexp.String())
	}
	for _, exp := range sexp.List[1:] {
		var poly PolyExp
		if err := poly.Parse(exp); err != nil {
			return fmt.Errorf("%s, failed to parse sub expression %s as polynomial while parsing sum exp %s", err, exp.String(), sexp.String())
		}
		sum.ps = append(sum.ps, poly)
	}
	return nil
}

type PolyExp struct {
	// Invariant: union type, at most one field allowed to be populated the other must be nil
	s *SumExp
	m *MonomialExp
	c *ConstantExp
	p *ProductExp
}

func (p *PolyExp) IsSum() bool {
	return p.s != nil
}

func (p *PolyExp) IsMon() bool {
	return p.m != nil
}

func (p *PolyExp) IsProduct() bool {
	return p.p != nil
}

func (p *PolyExp) IsConstant() bool {
	return p.c != nil
}

func (p *PolyExp) Sum() (*SumExp, error) {
	if p.s == nil {
		return nil, fmt.Errorf("polynomial is not a sum expression")
	}
	return p.s, nil
}

func (p *PolyExp) Mon() (*MonomialExp, error) {
	if p.m == nil {
		return nil, fmt.Errorf("polynomial is not a monomial expression")
	}
	return p.m, nil
}

func (p *PolyExp) check() error {
	var nilCount int
	if p.s == nil {
		nilCount++
	}
	if p.m == nil {
		nilCount++
	}
	if p.p == nil {
		nilCount++
	}
	if p.c == nil {
		nilCount++
	}

	if nilCount < 3 {
		return fmt.Errorf("overpopulated PolyExp")
	}
	if nilCount == 4 {
		return fmt.Errorf("unpopulated PolyExp")
	}
	return nil
}

// invariant polyexpression is valid
func (p *PolyExp) ToSExp() SExp {
	if p.IsMon() {
		return p.m.ToSExp()
	}
	if p.IsProduct() {
		return p.p.ToSExp()
	}
	if p.IsConstant() {
		return p.c.ToSExp()
	}

	return p.s.ToSExp()
}

// populate polynomial with contents of s-expression
func (p *PolyExp) Parse(sexp SExp) (err error) {
	defer func() {
		if err == nil {
			err = p.check()
		}
	}()

	// First check for constant exp
	if sexp.Atom != nil {
		var c ConstantExp
		if err := c.Parse(sexp); err != nil {
			return err
		}
		p.c = &c
		return nil
	}

	if len(sexp.List) < 1 {
		return fmt.Errorf("invalid SExp, cannot parse as polynomial %s", sexp.String())
	}
	var s SumExp
	var m MonomialExp
	var prod ProductExp

	if s.match(sexp.List[0]) {
		if err := s.Parse(sexp); err != nil {
			return err
		}
		p.s = &s
	}
	if m.match(sexp.List[0]) {
		if err := m.Parse(sexp); err != nil {
			return err
		}
		p.m = &m
	}
	if prod.match(sexp.List[0]) {
		if err := prod.Parse(sexp); err != nil {
			return err
		}
		p.p = &prod
	}

	return nil
}
