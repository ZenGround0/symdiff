package symdiff

import (
	"fmt"
	"strings"
	"strconv"
	"unicode"
)


/*
   TODOs
   1. Test S-Exp parsing
   2. Complete S-Exp -> PolyExp parsing
   3. PolyExp -> S-Exp serialization
   4. Differentiation function PolyExp -> PolyExp
   5. Test out e2e differentiation
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
   2. Colorful parentheses matching
   3. REPL -- diffierent commands
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
   - de next all sums into one flat sum expression
   - distribute products through all poly, sum, products and constants, only keep around monomials
   - add together all monomials of the same term
   - normalizze ( ^ x 0) to constant 1
   - drop zero constants   
   

*/

   

/*

   <poly exp>     ::= <sum exp> | <monomial exp>
   <sum exp>      ::= (sum <poly exp>...<poly exp>)
   <monomial exp> ::= (mon <int> <symbol> <int>)  // (mon a x n) == a(x^n)
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

func IsSymbol(s string) bool  {
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
const MonomialSugarKeyWord = "'"

// Valid atom strings that are not alphanumeric
var SpecialAtoms map[string]struct{}
var Rainbow []int

func init() {
	SpecialAtoms = make(map[string]struct{})
	SpecialAtoms[SumSugarKeyWord] = struct{}{}
	SpecialAtoms[MonomialSugarKeyWord] = struct{}{}

	Rainbow = make([]int, 6)
	Rainbow[1] = 124
	Rainbow[2] = 202
	Rainbow[3] = 11
	Rainbow[4] = 34
	Rainbow[5] = 51
	Rainbow[0] = 19
	
}

type MonomialExp struct {
	a int
	x Symbol
	n int
}

// Getter for all fields constituting monomial term
// Fields are private to restrict setting to parsing
func (m *MonomialExp) Term() (int, Symbol, int) {
	return m.a, m.x, m.n
}

// Match monomial to SExp head Atom
func (m *MonomialExp) match(s SExp) bool {
	if s.Atom == nil {
		return false
	}
	return *s.Atom == Atom(MonomialKeyWord) || *s.Atom == Atom(MonomialSugarKeyWord) 
}

func (m *MonomialExp) ToSExp() SExp {
	return SExp {
		List: []SExp{
			NewAtom("'"),
			NewAtom(fmt.Sprintf("%d", m.a)),
			NewAtom(string(m.x)),
			NewAtom(fmt.Sprintf("%d", m.n)),
		},
	}
}

func (m *MonomialExp) Parse(s SExp) error {
	if len(s.List) != 4 && !m.match(s.List[0]) { 
		return fmt.Errorf("invalid SExp, cannot parse as monomial %s", s.String())
	}
	for _, sexp := range s.List[1:] {
		if sexp.Atom == nil {
			return fmt.Errorf("Invalid SExp, cannot parse as monomial %s", s.String())
		}
	}
	a, err := strconv.Atoi(string(*s.List[1].Atom))
	if err != nil {
		return fmt.Errorf("%s failed to parse coefficient for monomial %s", err, s.String())
	}
	m.a = a
	if !IsSymbol(string(*s.List[2].Atom)) {
		return fmt.Errorf("failed to parse variable, not a valid symbol for monomial %s", s.String())
	}
	m.x = Symbol(*s.List[2].Atom)
	n, err := strconv.Atoi(string(*s.List[3].Atom))
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

	return SExp {
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
}

func (p *PolyExp) IsSum() bool {
	return p.s != nil 
}

func (p *PolyExp) IsMon() bool {
	return p.m != nil 
}

func (p *PolyExp) Sum() (*SumExp, error) {
	if p.s == nil {
		return nil, fmt.Errorf("Polynomial is not a sum expression")
	}
	return p.s, nil
}

func (p *PolyExp) Mon() (*MonomialExp, error) {
	if p.m == nil {
		return nil, fmt.Errorf("Polynomial is not a monomial expression")
	}
	return p.m, nil
}

func (p *PolyExp) check() error {
	if p.s == nil && p.m == nil {
		return fmt.Errorf("Unpopulated PolyExp")
	}
	if p.s != nil && p.m != nil {
		return fmt.Errorf("Overpopulated PolyExp")
	}
	return nil
}

func (p *PolyExp) match(sexp SExp) bool {
	var s SumExp
	var m MonomialExp
	return s.match(sexp) || m.match(sexp)
}

// invariant polyexpression is valid
func (p *PolyExp) ToSExp() SExp {
	if p.IsMon() {
		return p.m.ToSExp()
	}
	
	return p.s.ToSExp()
}

// populate polynomial with contents of s-expression
func (p *PolyExp) Parse(sexp SExp) (err error) {
	defer func () {
		if err == nil {
			err = p.check()
		}
	}()	
	if len(sexp.List) < 1 {
		return fmt.Errorf("invalid SExp, cannot parse as polynomial %s", sexp.String())
	}
	var s SumExp
	var m MonomialExp

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
	
	return nil
}
	

// datatypes for parsing S-expressions with strings as Atoms
type Atom string

type SExp struct {
	// Invariant: union type, at most one field allowed to be populated the other must be nil
	// both can be nil to match empty list ()
	Atom *Atom 
	List []SExp
}

func NewAtom(s string) SExp {
	a := new(Atom)
	*a = Atom(s)
	return SExp {
		Atom: a,
	}
}

func (s *SExp) check() error {
	if s.Atom != nil && s.List != nil {
		return fmt.Errorf("Overpopulated SExp")
	}
	return nil
}

func (s SExp) String() string {
	if s.Atom == nil && s.List == nil {
		return "( )"
	}
	
	if s.Atom != nil {
		return string(*s.Atom)
	}

	ret := "( "
	for _, sub := range s.List {
		ret += sub.String() + " " 
	}
	ret += ")"
	
	return ret
}

// Invariant: s and t internally consistent as verified with `check`
func (s *SExp) Equal(t *SExp) bool {
	// Empty list case
	if s.List == nil && s.Atom == nil {
		return t.List == nil && t.Atom == nil
	}
	
	// Atom case
	if s.List == nil {
		if t.Atom == nil {
			return false
		}
		return *s.Atom == *t.Atom
	}
	// List case
	if len(s.List) != len(t.List) {
		return false
	}
	for i := range s.List {
		if !s.List[i].Equal(&t.List[i]) {
			return false
		}
	}
	return true
}

// Parse an S-Expression out of a raw string
// Support for modern notation for succinct representation of lists with more than one element
// Limited atom support -- only alphanumeric strings and registered exceptions are allowed as atoms
/*
 "( A ( B 0) ())" parses to: 
 
 SExp{
    list: []SExp{
       SExp{
          atom: "A"
       },
       SExp{
          SExp{
             list: []SExp{
                SExp{
                   atom: "B"
                },
                SExp{
                   atom: "0"
                }
             }
          }
       },
       SExp{}
    }
 }
*/ 
func (s *SExp) Parse(raw string) (err error) {
	defer func () {
		if err == nil {
			err = s.check()
		}
	}()
	raw = strings.TrimSpace(raw)
	// atom case
	if isAtom(raw) {
		s.Atom = new(Atom)
		*s.Atom = Atom(raw)
		return nil
	}

	// list case
	listStr, err := Unwrap(raw)
	if err != nil {
		return err
	}
	expStr, remaining, err := TakeSExp(listStr)
	if err != nil {
		return err
	}
	for expStr != "" {
		var next SExp
		if err := next.Parse(expStr); err != nil {
			return err
		}
		s.List = append(s.List, next)
		expStr, remaining, err = TakeSExp(remaining)
		if err != nil {
			return err
		}
	}
	
	return nil
}


func isAtomChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
	
func isAtom (raw string) bool {
	if _, special := SpecialAtoms[raw]; special {
		return true
	}
	// excluding special atoms, atoms are contiguous alphanumeric strings excluding the empty string
	if len(raw) == 0 {
		return false
	}
	for _, r := range raw {
		if !isAtomChar(r) {
			return false
		}
	}
	return true
}


func Unwrap(raw string) (string, error) {
	raw = strings.TrimSpace(raw)	
	if len(raw) < 2 {
		return "", fmt.Errorf("bad S-Expression \"%s\", not enough chars to unwrap", raw)
	}
	if string(raw[0]) != "(" {
		return "", fmt.Errorf("bad S-Expression %s, no opening paren", raw)
	}
	if string(raw[len(raw)-1]) != ")" {
		return "", fmt.Errorf("bad S-Expression %s, no closing paren", raw)
	}
	return raw[1:len(raw)-1], nil
}

// Given unwrapped list SExp, take the next inner SExp.
// Return next and remaining.  When next == "" traversal is done
func TakeSExp(raw string) (string, string, error) {
	// add whitespace characters around all parenthesis to support the common ( atom(..) ) and ( () atom) cases
	raw = strings.Replace(raw, "(", " ( ", -1)
	raw = strings.Replace(raw, ")", " ) ", -1)	
	
	sExpList := strings.TrimSpace(raw)
	// raw, _  = Unwrap("( )")
	if len(sExpList) == 0 {
		return "", "", nil
	}

	// raw, _ = Unwrap("( x ... )")
	if string(sExpList[0]) != "(" {
		elts := strings.Fields(sExpList)
		if len(elts) == 0 {
			return "", "", fmt.Errorf("internal error parsing sExp list %s, expected non zero sExps in list", raw)
		}
		if !isAtom(elts[0]) {
			return "", "", fmt.Errorf("invalid SExp list %s, invalid atom at head of list", raw)
		}
		return elts[0], strings.Join(elts[1:], " "), nil
	}

	// raw, _ = Unwrap("( (...) ...)")
	// find closing paren
	open := 0
	var end int
	for i, r := range sExpList {
		if string(r) == "(" {
			open += 1
		}
		if string(r) == ")" {
			open -= 1
			if open == 0 {
				end = i
				break
			}
		}
	}
	if open != 0 {
		return "", "", fmt.Errorf("invalid SExp list %s, mismatching parens in SExp at head of list", raw)
	}
	return sExpList[0:end+1], sExpList[end+1:], nil
}


// Take a string with repeated spaces and replace them with single spaces
func NormalizeSpaces(s string) string {
	for strings.Contains(s, "  ") {
		s = strings.Replace(s, "  ", " ", -1)
	}
	return s
}


// Invariant: s has no mismatched parentheses
func RainbowParens(s string, rainbow []int) (string, error) {
	if len(rainbow) == 0 {
		return "", fmt.Errorf("Need to specify colors")
	}

	// traverse string subbing out parens for sequences coloring parens
	// use an implict stack by keeping a counter of open parens

	var out string
	open := 0
	for _, r := range s {
		if r == '(' {
			open++
			idx := open % len(rainbow)
			rbv := rainbow[idx]
			out += fmt.Sprintf("\033[38;5;%dm(\033[0m", rbv)
		} else if r == ')' {
			idx := open % len(rainbow)
			rbv := rainbow[idx]
			out += fmt.Sprintf("\033[38;5;%dm)\033[0m", rbv)
			if open == 0 {
				return "", fmt.Errorf("mismatched parentheses")
			}
			open--
		} else {
			out += string(r) 
		}
	}
	if open > 0 {
		return "", fmt.Errorf("mismatched parentheses")
	}
	return out, nil
}
