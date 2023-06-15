package symdiff_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/zenground0/symdiff"
)

func TestSimplifyFlattenAndJoin(t *testing.T) {

	// nested sums
	poly := polyFromString(t, "(+ 1 ( + (^ x 1) ( + (^ x 2) ( + (^ x 3) (^ x 4))))) ")

	flattened := Flatten(poly)
	assert.Equal(t, 5, len(flattened))
	assert.Equal(t, "( ^ x 2 )", flattened[2].ToSExp().String())
	joined := Join(flattened)
	assert.Equal(t, "( + 1 ( ^ x 1 ) ( ^ x 2 ) ( ^ x 3 ) ( ^ x 4 ) )", joined.ToSExp().String())
}

func TestFlattenSkipsProducts(t *testing.T) {
	poly := polyFromString(t, "(+ 1 ( + (^ x 1) ( + (^ x 2) ( + (* 3 ( * 3 ( * 3 (^ x 3) )  ) ) (^ x 4))))) ")
	flattened := Flatten(poly)
	assert.Equal(t, 5, len(flattened))

	assert.Equal(t, "( * 3 ( * 3 ( * 3 ( ^ x 3 ) ) ) )", flattened[3].ToSExp().String())
	joined := Join(flattened)
	assert.Equal(t, "( + 1 ( ^ x 1 ) ( ^ x 2 ) ( * 3 ( * 3 ( * 3 ( ^ x 3 ) ) ) ) ( ^ x 4 ) )", joined.ToSExp().String())
}

func TestApplyProducts(t *testing.T) {
	polyConst := polyFromString(t, "( * 3 ( * 3 3 ) )")
	polyProd, err := ApplyProducts(1, polyConst)
	assert.NoError(t, err)
	assert.Equal(t, "27", polyProd.ToSExp().String())

	polyDist := polyFromString(t, "( * 3 ( + ( ^ x 0 ) ( ^ x 0 )))")
	polyProd, err = ApplyProducts(1, polyDist)
	assert.NoError(t, err)
	assert.Equal(t, "( + ( * 3 ( ^ x 0 ) ) ( * 3 ( ^ x 0 ) ) )", polyProd.ToSExp().String())

	polyTwoLayers := polyFromString(t, "( * 3 ( + ( * 2 ( ^ x 0 ) ) ( * 5 ( ^ x 0 ) ) ) )")
	polyProd, err = ApplyProducts(1, polyTwoLayers)
	assert.NoError(t, err)
	assert.Equal(t, "( + ( * 6 ( ^ x 0 ) ) ( * 15 ( ^ x 0 ) ) )", polyProd.ToSExp().String())
}

func TestFold(t *testing.T) {
	poly := polyFromString(t, "( + ( ^ x 2) ( * 2 ( ^ x 2 ) ) )")
	polyFold := Join(Fold(Flatten(poly)))
	assert.Equal(t, "( + ( * 3 ( ^ x 2 ) ) 0 )", polyFold.ToSExp().String())

	poly = polyFromString(t, "( + 1 2 3 )")
	polyFold = Join(Fold(Flatten(poly)))
	assert.Equal(t, "6", polyFold.ToSExp().String())

	// nested products ignored
	poly = polyFromString(t, "( + ( ^ x 2) ( * 2 ( ^ x 2 ) ) ( * 2 ( * 1 ( ^ x 2 ) ) ) )")
	polyFold = Join(Fold(Flatten(poly)))
	assert.Equal(t, "( + ( * 2 ( * 1 ( ^ x 2 ) ) ) ( * 3 ( ^ x 2 ) ) 0 )", polyFold.ToSExp().String())

	// x^0 is constant
	poly = polyFromString(t, "( + 1 2 3 ( * 4 (^ x 0 ) ) )")
	polyFold = Join(Fold(Flatten(poly)))
	assert.Equal(t, "10", polyFold.ToSExp().String())
}
