package symdiff

import (
	"fmt"
	"strings"
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
   
   Extension ideas-- 
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
   2. Chain rule
      - d/dx (f(u)) = df/dx * du/dx
      - prep: think about how to modify expressions + write out a test case
      - goal: add composite polys to S-exp, differentiate fn, test in repl
         - i.e. ((x + 5)^7)^-1  --> 
   3. Product rule
      - d/dx (f * g) = df/dx
      - prep: think about how to modify expressions + write out a test case
      - goal: add products to S-exp, differentiate fn, test in repl
        - i.e. (x + 5) * (x^7)  --> (1)x^7 + (x +5)*7x^6
   3.5 Rational Functions
      - we should do a specialized version of chain rule for power expressions
      - do this plus product rule then we get rational functions and that's cool
      - good goal to shoot for
   4. Infinite sums
      - (+inf n (' 1 x n) ), bind a summation variable n and create a term
      - derivative treams n as a constant
   5. Exponential from taylor series
      - need to deal with 1 / n!
      - renormalize n-1 to n 
   
*/

/*

   <poly exp>     ::= <sum exp> | <monomial exp>
   <sum exp>      ::= (sum <poly exp>...<poly exp>)
   <monomial exp> ::= (mon <int> <symbol> <int>)  // (mon a x n) == a(x^n)
   <symbol>       ::= alphabetical string
   <int>          ::= integer string 

Syntactic sugar: 
sum := s | + 
mon := m | ‘ 

Example

   `5 x^5 + 6 x ^3 - x + 6 + 3 x^-7`
   (+ (‘ 5 x 5) (‘ 6 x 3 ) (‘ -1 x 1) (‘ 6 x 0) (‘ 3 x -7  ))

   Derivative should be 

   `25 x ^4 + 18 x ^ 2 - 1 + 0 - 21 x ^-8`

   (+ (‘ 25 x 4) (‘ 18 x 2) (‘ -1 x 0) (‘ -21 x -8))

*/


// datatypes for defining internal expressions of polynomial functions
type Symbol struct {
	s string
}

const SumKeyWord = "sum"
const SumSugarKeyWord = "+"
const MonomialKeyWord = "mon"
const MonomialSugarKeyWord = "'"

type MonomialExp struct {
	a int
	x Symbol
	n int
}

type SumExp struct {
	m MonomialExp
	p PolyExp
}

func (s *SumExp) match(sexp SExp) bool {
	//	if sexp.start ==
	return false
}	

type PolyExp struct {
	// Invariant: union type, at most one field allowed to be populated the other must be nil
	s *SumExp
	m *MonomialExp
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

// populate polynomial with contents of s-expression
func (p *PolyExp) Parse(sexp SExp) error {
	//	if sexp.start ==
	return nil
}
	

// datatypes for parsing S-expressions with strings as Atoms
type Atom string

type SExp struct {
	// Invariant: union type, at most one field allowed to be populated the other must be nil
	// both can be nil to match empty list ()
	atom *Atom 
	list []SExp
}

func (s *SExp) check() error {
	if s.atom != nil && s.list != nil {
		return fmt.Errorf("Overpopulated SExp")
	}
	return nil
}

// Parse an S-Expression out of a raw string
// Support for modern notation for succinct representation of lists with more than one element
// Limited atom support -- only alphanumeric strings are allowed as atoms
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
	isAtom := func(raw string) bool {
		// atoms are contiguous alphanumeric strings excluding the empty string
		if len(raw) == 0 {
			return false
		}
		for _, r := range raw {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return false
			}
		}
		return true
	}

	unwrap := func(raw string) (string, error) {
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

	// atom case
	if isAtom(raw) {
		s.atom = new(Atom)
		*s.atom = Atom(raw)
		return nil
	}

	// list case
	listStr, err := unwrap(raw)
	if err != nil {
		return err
	}
	subExpStrs := strings.Fields(listStr)
	for _, expStr := range subExpStrs {
		var next SExp
		if err := next.Parse(expStr); err != nil {
			return err
		}
		s.list = append(s.list, next)
	}
	return nil
}

