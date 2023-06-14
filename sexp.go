package symdiff

import (
	"fmt"
	"strings"
	"unicode"
)

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
	return SExp{
		Atom: a,
	}
}

func (s *SExp) check() error {
	if s.Atom != nil && s.List != nil {
		return fmt.Errorf("overpopulated SExp")
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
	defer func() {
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
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-'
}

func isAtom(raw string) bool {
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
	return raw[1 : len(raw)-1], nil
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
	return sExpList[0 : end+1], sExpList[end+1:], nil
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
		return "", fmt.Errorf("need to specify colors")
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
