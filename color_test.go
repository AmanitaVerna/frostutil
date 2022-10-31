package frostutil

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UnmultiplyAlpha(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 100}
	r, g, b, a := UnmultiplyAlpha(c)
	assert.Equal(t, byte(255), r)
	assert.Equal(t, byte(255), g)
	assert.Equal(t, byte(255), b)
	assert.Equal(t, byte(100), a)

	c = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	r, g, b, a = UnmultiplyAlpha(c)
	assert.Equal(t, byte(100), r)
	assert.Equal(t, byte(100), g)
	assert.Equal(t, byte(100), b)
	assert.Equal(t, byte(255), a)

	c = color.RGBA{R: 17, G: 51, B: 68, A: 85}
	r, g, b, a = UnmultiplyAlpha(c)
	assert.Equal(t, byte(51), r)
	assert.Equal(t, byte(153), g)
	assert.Equal(t, byte(204), b)
	assert.Equal(t, byte(85), a)
}

func Test_UnmultiplyAlphaBytes(t *testing.T) {
	r, g, b, a := UnmultiplyAlphaBytes(100, 100, 100, 100)
	assert.Equal(t, byte(255), r)
	assert.Equal(t, byte(255), g)
	assert.Equal(t, byte(255), b)
	assert.Equal(t, byte(100), a)

	r, g, b, a = UnmultiplyAlphaBytes(100, 100, 100, 255)
	assert.Equal(t, byte(100), r)
	assert.Equal(t, byte(100), g)
	assert.Equal(t, byte(100), b)
	assert.Equal(t, byte(255), a)

	r, g, b, a = UnmultiplyAlphaBytes(17, 51, 68, 85)
	assert.Equal(t, byte(51), r)
	assert.Equal(t, byte(153), g)
	assert.Equal(t, byte(204), b)
	assert.Equal(t, byte(85), a)
}

func Test_MultiplyAlphaBytes(t *testing.T) {
	r, g, b, a := MultiplyAlphaBytes(255, 255, 255, 100)
	assert.Equal(t, byte(100), r)
	assert.Equal(t, byte(100), g)
	assert.Equal(t, byte(100), b)
	assert.Equal(t, byte(100), a)

	r, g, b, a = MultiplyAlphaBytes(100, 100, 100, 255)
	assert.Equal(t, byte(100), r)
	assert.Equal(t, byte(100), g)
	assert.Equal(t, byte(100), b)
	assert.Equal(t, byte(255), a)

	r, g, b, a = MultiplyAlphaBytes(51, 153, 204, 85)
	assert.Equal(t, byte(17), r)
	assert.Equal(t, byte(51), g)
	assert.Equal(t, byte(68), b)
	assert.Equal(t, byte(85), a)

	r, g, b, a = MultiplyAlphaBytes(42, 87, 197, 0)
	assert.Equal(t, byte(0), r)
	assert.Equal(t, byte(0), g)
	assert.Equal(t, byte(0), b)
	assert.Equal(t, byte(0), a)
}

func Test_MultiplyAlphaBytesPreserveColors(t *testing.T) {
	r, g, b, a := MultiplyAlphaBytesPreserveColors(255, 255, 255, 100)
	assert.Equal(t, byte(100), r)
	assert.Equal(t, byte(100), g)
	assert.Equal(t, byte(100), b)
	assert.Equal(t, byte(100), a)

	r, g, b, a = MultiplyAlphaBytesPreserveColors(100, 100, 100, 255)
	assert.Equal(t, byte(100), r)
	assert.Equal(t, byte(100), g)
	assert.Equal(t, byte(100), b)
	assert.Equal(t, byte(255), a)

	r, g, b, a = MultiplyAlphaBytesPreserveColors(51, 153, 204, 85)
	assert.Equal(t, byte(17), r)
	assert.Equal(t, byte(51), g)
	assert.Equal(t, byte(68), b)
	assert.Equal(t, byte(85), a)

	r, g, b, a = MultiplyAlphaBytesPreserveColors(42, 87, 197, 0)
	assert.Equal(t, byte(42), r)
	assert.Equal(t, byte(87), g)
	assert.Equal(t, byte(197), b)
	assert.Equal(t, byte(0), a)
}

func Test_ToNRGBA(t *testing.T) {
	c := color.RGBA{R: 17, G: 51, B: 68, A: 85}
	r, g, b, a := ToNRGBA(c)
	assert.Equal(t, byte(51), r)
	assert.Equal(t, byte(153), g)
	assert.Equal(t, byte(204), b)
	assert.Equal(t, byte(85), a)
}

func Test_ToNRGBA_U32(t *testing.T) {
	c := color.NRGBA{R: 0xde, G: 0xad, B: 0xbe, A: 0xef}
	u := ToNRGBA_U32(c)
	assert.EqualValues(t, 0xdeadbeef, u)
}
