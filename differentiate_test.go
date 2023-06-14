package symdiff_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "github.com/zenground0/symdiff"
)

func TestDiffConstant(t *testing.T) {
	poly := polyFromString(t, "70")

	derivative, err := Differentiate("x", poly)
	assert.NoError(t, err)
	assert.Equal(t, Zero(), *derivative)
}

func TestDiffMonomial(t *testing.T) {
	poly := polyFromString(t, "(^ x 8)")

	derivative, err := Differentiate("x", poly)
	assert.NoError(t, err)

	expected := "( * 8 ( ^ x 7 ) )"
	assert.Equal(t, expected, derivative.ToSExp().String())
}

func TestDiffProduct(t *testing.T) {
	poly := polyFromString(t, "(* 7 (^ x 2))")

	derivative, err := Differentiate("x", poly)
	assert.NoError(t, err)
	expected := "( * 7 ( * 2 ( ^ x 1 ) ) )"
	assert.Equal(t, expected, derivative.ToSExp().String())
}

func TestDiffSumm(t *testing.T) {
	poly := polyFromString(t, "( + (* 7 (^ x 2)) (^ x 8) )")

	derivative, err := Differentiate("x", poly)
	assert.NoError(t, err)
	expected := "( + ( * 7 ( * 2 ( ^ x 1 ) ) ) ( * 8 ( ^ x 7 ) ) )"
	assert.Equal(t, expected, derivative.ToSExp().String())
}

func polyFromString(t *testing.T, raw string) PolyExp {
	var sexp SExp
	require.NoError(t, sexp.Parse(raw))
	var poly PolyExp
	require.NoError(t, poly.Parse(sexp))
	return poly
}
