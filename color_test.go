package frostutil

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UnmultiplyAlpha(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 100}
	r, g, b, a := unmultiplyAlpha(c)

	assert.Equal(t, 255, r)
	assert.Equal(t, 255, g)
	assert.Equal(t, 255, b)
	assert.Equal(t, 100, a)

	c = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	r, g, b, a = unmultiplyAlpha(c)

	assert.Equal(t, 100, r)
	assert.Equal(t, 100, g)
	assert.Equal(t, 100, b)
	assert.Equal(t, 255, a)
}

func Test_ToNRGBA(t *testing.T) {
	c := color.RGBA{R: 17, G: 51, B: 68, A: 85}
	r, g, b, a := ToNRGBA(c)
	assert.Equal(t, 51, r)
	assert.Equal(t, 153, g)
	assert.Equal(t, 204, b)
	assert.Equal(t, 85, a)
}

func Test_ToNRGBA_U32(t *testing.T) {
	c := color.NRGBA{R: 0xde, G: 0xad, B: 0xbe, A: 0xef}
	u := ToNRGBA_U32(c)
	assert.EqualValues(t, 0xdeadbeef, u)
}
