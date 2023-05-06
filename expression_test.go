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

func TestParseSExpDeep(t *testing.T) {
	raw := "( + ( + ( + ( + 1 2 ) 3 ) 4 ) 5 )"
	var sexp SExp
	assert.NoError(t, sexp.Parse(raw))
	assert.True(t, len(sexp.List) == 3)
	s := sexp
	for i := 0; i < 3; i ++ {
		assert.True(t, len(s.List) == 3)
		assert.True(t, s.List[0].List == nil) // +
		assert.True(t, s.List[2].List == nil) // n	
		s = s.List[1]
	}
		
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
	assert.Equal(t, "( + ( m 3 x 6 ) ( ' 5 x 0 ) )", packedSExp.String())
	assert.Equal(t, "( + ( m 3 x 6 ) ( ' 5 x 0 ) )", spacedSExp.String())
}

func polyMon(t *testing.T, poly PolyExp) *MonomialExp {
	m, err := poly.Mon()
	require.NoError(t, err)
	return m
}

func polySum(t *testing.T, poly PolyExp) *SumExp {
	s, err := poly.Sum()
	require.NoError(t, err)
	return s
}

func TestParsePolynomials(t *testing.T) {
	var sexp SExp
	assert.NoError(t, sexp.Parse("( ' 1 x 2 )"), "sexp parse error")
	var poly PolyExp
	assert.NoError(t, poly.Parse(sexp), "polynomial parse error")
	assert.True(t, poly.IsMon())
	a, x, n := polyMon(t, poly).Term()
	assert.True(t, n == 2 && a == 1 && x == Symbol("x"))
	assert.Equal(t, "( ' 1 x 2 )", poly.ToSExp().String())

	var sexp2 SExp
	assert.NoError(t, sexp2.Parse("( + ( + ( mon 1 x 0) ( mon 1 y 1 ) ) ( ' 1 x 2 ) )"))
	var poly2 PolyExp
	assert.NoError(t, poly2.Parse(sexp2), "polynomial parse error")
	assert.Equal(t, "( + ( + ( ' 1 x 0 ) ( ' 1 y 1 ) ) ( ' 1 x 2 ) )", poly2.ToSExp().String())	
	assert.True(t, poly2.IsSum())
	ps := polySum(t, poly2).Term()
	require.Len(t, ps, 2)
	polyFirst, polySecond := ps[0], ps[1]
	assert.True(t, polyFirst.IsSum() && polySecond.IsMon())
	a, x, n = polyMon(t, polySecond).Term()
	assert.True(t, n == 2, a == 1, x == Symbol("x"))
	assert.Equal(t, "( ' 1 x 2 )", polySecond.ToSExp().String())
	ps = polySum(t, polyFirst).Term()
	require.Len(t, ps, 2)
	assert.Equal(t, "( + ( ' 1 x 0 ) ( ' 1 y 1 ) )", polyFirst.ToSExp().String())	
	assert.True(t, ps[0].IsMon(), ps[1].IsMon())
	a, x, n = polyMon(t, ps[0]).Term()
	assert.True(t, n == 0, a == 1, x == Symbol("x"))
	assert.Equal(t, "( ' 1 x 0 )", ps[0].ToSExp().String())
	a, x, n = polyMon(t, ps[1]).Term()
	assert.True(t, n == 1, a == 1, x == Symbol("y"))
	assert.Equal(t, "( ' 1 y 1 )", ps[1].ToSExp().String())	
}
