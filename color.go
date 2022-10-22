package frostutil

import (
	"image/color"
)

// ToNRGBA converts a color to RGBA values which are not premultiplied, unlike color.RGBA().
func ToNRGBA(c color.Color) (r, g, b, a int) {
	// We use UnmultiplyAlpha with RGBA, RGBA64, and unrecognized implementations of Color.
	// It works for all Colors whose RGBA() method is implemented according to spec, but is only necessary for those.
	// Only RGBA and RGBA64 have components which are already premultiplied.
	switch col := c.(type) {
	// NRGBA and NRGBA64 are not premultiplied
	case color.NRGBA:
		r = int(col.R)
		g = int(col.G)
		b = int(col.B)
		a = int(col.A)
	case *color.NRGBA:
		r = int(col.R)
		g = int(col.G)
		b = int(col.B)
		a = int(col.A)
	case color.NRGBA64:
		r = int(col.R) >> 8
		g = int(col.G) >> 8
		b = int(col.B) >> 8
		a = int(col.A) >> 8
	case *color.NRGBA64:
		r = int(col.R) >> 8
		g = int(col.G) >> 8
		b = int(col.B) >> 8
		a = int(col.A) >> 8
	// Gray and Gray16 have no alpha component
	case *color.Gray:
		r = int(col.Y)
		g = int(col.Y)
		b = int(col.Y)
		a = 0xff
	case color.Gray:
		r = int(col.Y)
		g = int(col.Y)
		b = int(col.Y)
		a = 0xff
	case *color.Gray16:
		r = int(col.Y) >> 8
		g = int(col.Y) >> 8
		b = int(col.Y) >> 8
		a = 0xff
	case color.Gray16:
		r = int(col.Y) >> 8
		g = int(col.Y) >> 8
		b = int(col.Y) >> 8
		a = 0xff
	// Alpha and Alpha16 contain only an alpha component.
	case color.Alpha:
		r = 0xff
		g = 0xff
		b = 0xff
		a = int(col.A)
	case *color.Alpha:
		r = 0xff
		g = 0xff
		b = 0xff
		a = int(col.A)
	case color.Alpha16:
		r = 0xff
		g = 0xff
		b = 0xff
		a = int(col.A) >> 8
	case *color.Alpha16:
		r = 0xff
		g = 0xff
		b = 0xff
		a = int(col.A) >> 8
	default: // RGBA, RGBA64, and unknown implementations of Color
		r, g, b, a = unmultiplyAlpha(c)
	}
	return
}

// ToNRGBA_Color runs c through ToNRGBA and then packages its output into a color.NRGBA, which it returns.
func ToNRGBA_Color(c color.Color) (out color.Color) {
	cr, cg, cb, ca := ToNRGBA(c)
	out = color.NRGBA{R: byte(cr), G: byte(cg), B: byte(cb), A: byte(ca)}
	return
}

// ToNRGBA_U32 converts a color to a uint32 where the highest byte is red, the next highest is green, the third highest is blue, and the lowest byte is alpha.
func ToNRGBA_U32(c color.Color) (u uint32) {
	cr, cg, cb, ca := ToNRGBA(c)
	u = (uint32(cr) << 24) | (uint32(cg) << 16) | (uint32(cb) << 8) | (uint32(ca))
	return
}

// unmultiplyAlpha returns a color's RGBA components as 8-bit integers by calling c.RGBA() and then removing the alpha premultiplication.
// It is only used by ToRGBA.
func unmultiplyAlpha(c color.Color) (r, g, b, a int) {
	red, green, blue, alpha := c.RGBA()
	if alpha != 0 && alpha != 0xffff {
		red = (red * 0xffff) / alpha
		green = (green * 0xffff) / alpha
		blue = (blue * 0xffff) / alpha
	}
	// Convert from range 0-65535 to range 0-255
	r = int(red >> 8)
	g = int(green >> 8)
	b = int(blue >> 8)
	a = int(alpha >> 8)
	return
}

// UnmultiplyAlphaBytes converts RGBA bytes to NRGBA bytes by removing the alpha premultiplication.
// It's for use with directly converting bytes in an image's Pix buffer.
func UnmultiplyAlphaBytes(red, green, blue, alpha byte) (r, g, b, a byte) {
	// If alpha is 0 or 0xff, we don't need to unmultiply.
	if alpha != 0 && alpha != 0xff {
		red = byte((int(red) * 0xff) / int(alpha))
		green = byte((int(green) * 0xff) / int(alpha))
		blue = byte((int(blue) * 0xff) / int(alpha))
	}
	// copy the values from the input vars to the output vars.
	r = red
	g = green
	b = blue
	a = alpha
	return
}
