package symdiff_test

import (
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"	
	. "github.com/zenground0/symdiff"
)


func TestParsingHelpers(t *testing.T) {
	raw := "   (   A   (B 0 ) () (   ) ) "
	start, err := Unwrap(raw)
	assert.NoError(t, err)
	next, remaining, err := TakeSExp(start)
	assert.NoError(t, err)
	assert.Equal(t, "A", next)
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "( B 0 )", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "( )", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "( )", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "", next)
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "", next)		
}

// Test parsing SExps of the form we care about
func TestParsingHelpersPoly(t *testing.T) {
	raw := " (   + ( ' 1 x 1) ( ' 4 x 2) ( + ( ' 3 x 0 ) ( ' 5 x 0 ) )) "

	start, err := Unwrap(raw)
	assert.NoError(t, err)
	next, remaining, err := TakeSExp(start)
	assert.NoError(t, err)
	assert.Equal(t, "+", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "( ' 1 x 1 )", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "( ' 4 x 2 )", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "( + ( ' 3 x 0 ) ( ' 5 x 0 ) )", NormalizeSpaces(next))

	start, err = Unwrap(next)
	assert.NoError(t, err)
	next, remaining, err = TakeSExp(start)
	assert.NoError(t, err)
	assert.Equal(t, "+", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "( ' 3 x 0 )", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "( ' 5 x 0 )", NormalizeSpaces(next))

	start, err = Unwrap(next)
	assert.NoError(t, err)
	next, remaining, err = TakeSExp(start)
	assert.NoError(t, err)
	assert.Equal(t, "'", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "5", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "x", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "0", NormalizeSpaces(next))
	next, remaining, err = TakeSExp(remaining)
	assert.NoError(t, err)
	assert.Equal(t, "", NormalizeSpaces(next))	
}

func TestParseSExp(t *testing.T) {
	raw := "( A ( B 0) ())"
	a, b, zero := new(Atom), new(Atom), new(Atom)
	*a, *b, *zero = "A", "B", "0"
	expected := SExp{
		List: []SExp{
			SExp{
				Atom: a,
				List: nil,
			},
			SExp{
				List: []SExp{
					SExp{
						Atom: b,
						List: nil,
					},
					SExp{
						Atom: zero,
						List: nil,
					},
				},
				Atom: nil,
			},
			SExp{
				Atom: nil,
				List: nil,
			},
		},
	}
	var observed SExp
	assert.NoError(t, observed.Parse(raw), "parse error")
	assert.True(t, expected.Equal(&observed))
}


func TestParseEmptySpaces(t *testing.T) {
	packed := "(+(m 3 x 6)(' 5 x 0))"
	spaced := fmt.Sprintf("  (+   \t\t\t\n( m             3 x 6   )     (' 5 x 0\r)\n\n)")
	var packedSExp SExp
	var spacedSExp SExp
	assert.NoError(t, packedSExp.Parse(packed), "parse error")
	assert.NoError(t, spacedSExp.Parse(spaced), "parse error")	
}

func polyMon(t *testing.T, poly PolyExp) MonomialExp {
	m, err := poly.Mon()
	require.NoError(t, err)
	return m
}

func TestParsePolynomials(t *testing.T) {
	var sexp SExp
	assert.NoError(t, sexp.Parse("( ' 1 x 2 )"), "sexp parse error")
	var poly PolyExp
	assert.NoError(t, poly.Parse(sexp), "polynomial parse error")
	assert.True(t, poly.IsMon())
	a, x, n := polyMon(t, poly).Term()
	assert.True(t, n == 2 && a == 1 && x == 'x')
}
