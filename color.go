package frostutil

import (
	"image/color"
)

// ToNRGBA converts a color to 8-bit RGBA values which are not premultiplied, unlike color.RGBA().
// This has special fast code for color.NRGBA, color.NRGBA64, color.Gray, color.Gray16, color.Alpha, and color.Alpha16, since none of those are premultiplied.
// For RGBA and RGBA64, it calls our UnmultiplyAlpha function, which both un-premultiplies the alpha from the RGB components, and reduces the color to 8bpp.
// UnmultiplyAlpha only un-premultiplies when the alpha returned by c.RGBA() is > 0 and < 0xffff.
func ToNRGBA(c color.Color) (r, g, b, a byte) {
	// We use UnmultiplyAlpha with RGBA, RGBA64, and unrecognized implementations of Color.
	// It works for all Colors whose RGBA() method is implemented according to spec, but is only necessary for those.
	// Only RGBA and RGBA64 have components which are already premultiplied.
	switch col := c.(type) {
	// NRGBA and NRGBA64 are not premultiplied
	case color.NRGBA:
		r = col.R
		g = col.G
		b = col.B
		a = col.A
	case *color.NRGBA:
		r = col.R
		g = col.G
		b = col.B
		a = col.A
	case color.NRGBA64:
		r = byte(col.R >> 8)
		g = byte(col.G >> 8)
		b = byte(col.B >> 8)
		a = byte(col.A >> 8)
	case *color.NRGBA64:
		r = byte(col.R >> 8)
		g = byte(col.G >> 8)
		b = byte(col.B >> 8)
		a = byte(col.A >> 8)
	// Gray and Gray16 have no alpha component
	case *color.Gray:
		r = col.Y
		g = col.Y
		b = col.Y
		a = 0xff
	case color.Gray:
		r = col.Y
		g = col.Y
		b = col.Y
		a = 0xff
	case *color.Gray16:
		r = byte(col.Y >> 8)
		g = byte(col.Y >> 8)
		b = byte(col.Y >> 8)
		a = 0xff
	case color.Gray16:
		r = byte(col.Y >> 8)
		g = byte(col.Y >> 8)
		b = byte(col.Y >> 8)
		a = 0xff
	// Alpha and Alpha16 contain only an alpha component.
	case color.Alpha:
		r = 0xff
		g = 0xff
		b = 0xff
		a = col.A
	case *color.Alpha:
		r = 0xff
		g = 0xff
		b = 0xff
		a = col.A
	case color.Alpha16:
		r = 0xff
		g = 0xff
		b = 0xff
		a = byte(col.A >> 8)
	case *color.Alpha16:
		r = 0xff
		g = 0xff
		b = 0xff
		a = byte(col.A >> 8)
	default: // RGBA, RGBA64, and unknown implementations of Color
		r, g, b, a = UnmultiplyAlpha(c)
	}
	return
}

// ToNRGBA_Color runs c through ToNRGBA and then packages its output into a color.NRGBA, which it returns.
func ToNRGBA_Color(c color.Color) (out color.Color) {
	cr, cg, cb, ca := ToNRGBA(c)
	out = color.NRGBA{R: byte(cr), G: byte(cg), B: byte(cb), A: byte(ca)}
	return
}

// ToNRGBA_U32 runs c through ToNRGBA and then packages its output into a uint32 where the highest byte is red, the next highest is green,
// the third highest is blue, and the lowest byte is alpha.
func ToNRGBA_U32(c color.Color) (u uint32) {
	cr, cg, cb, ca := ToNRGBA(c)
	u = (uint32(cr) << 24) | (uint32(cg) << 16) | (uint32(cb) << 8) | (uint32(ca))
	return
}

// UnmultiplyAlpha returns a color's RGBA components as 8-bit integers by calling c.RGBA() and then removing the alpha premultiplication (if present),
// and finally bitshifting each component right by 8 (>> 8) to reduce it from the 16-bit component output of RGBA() to 8-bit component output.
// The un-premultiplication is skipped if the alpha returned by c.RGBA() is 0 or 0xffff.
// This preserves non-zero color components when the alpha value is zero, unlike color.NRGBAModel.Convert.
// The standard model conversion function and NRGBA colors' RGBA() method erase the color information when alpha is zero, so this capability
// is only relevant if you are manually editing the image's pixel buffer. If you want to preserve color information when encoding from NRGBA to RGBA, you can use
// MultiplyAlphaBytesPreserveColors.
func UnmultiplyAlpha(c color.Color) (r, g, b, a byte) {
	red, green, blue, alpha := c.RGBA()
	a = byte(alpha >> 8)
	if alpha != 0 && alpha != 0xffff {
		// NRGBA.RGBA() returns red = rr * a / 0xff, same for green and blue but with gg and bb instead of rr, and returns alpha = aa, where xx means (x | (x << 8))
		// To reverse this, we can do:
		// a = alpha >> 8 (or alpha & 0xff, but we do >> 8 because of things like RGBA16 where the low byte and high byte won't be equal)
		// r = (red * 0xff / a) >> 8
		// etc
		r = byte(((red * 0xff) / uint32(a)) >> 8)
		g = byte(((green * 0xff) / uint32(a)) >> 8)
		b = byte(((blue * 0xff) / uint32(a)) >> 8)
	} else {
		// Convert from range 0-65535 to range 0-255
		r = byte(red >> 8)
		g = byte(green >> 8)
		b = byte(blue >> 8)
	}
	return
}

// UnmultiplyAlphaBytes converts alpha-premultiplied RGBA bytes to NRGBA bytes by removing the alpha premultiplication.
// It's for use when directly converting bytes in an image's Pix buffer.
// This preserves non-zero color components when the alpha component is zero, unlike color.NRGBAModel.Convert.
// The standard model conversion function and NRGBA colors' RGBA() method erase the color information when alpha is zero, so this capability
// is only relevant if you are manually editing the image's pixel buffer. If you want to preserve color information when encoding from NRGBA to RGBA, you can use
// MultiplyAlphaBytesPreserveColors.
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

// MultiplyAlphaBytes converts NRGBA bytes to RGBA bytes.
// Currently, it just packs the bytes into a color.NRGBA, calls RGBA() on it, and right-bitshifts each returned uint32 by 8, returning them as bytes
// (since RGBA() returns 16-bit numbers in uint32s).
// Note: Because this calls RGBA(), it loses the color components when alpha is zero.
func MultiplyAlphaBytes(red, green, blue, alpha byte) (r, g, b, a byte) {
	col := color.NRGBA{R: red, G: green, B: blue, A: alpha}
	ir, ig, ib, ia := col.RGBA() // get alpha-premultiplied rgba values
	r = byte(ir >> 8)
	g = byte(ig >> 8)
	b = byte(ib >> 8)
	a = byte(ia >> 8)
	return
}

// MultiplyAlphaBytesPreserveColors converts NRGBA bytes to RGBA bytes, preserving the color components when the alpha component is zero.
// Because it preserves the color components, it gives different results from the standard color model conversion methods and from calling RGBA() on an NRGBA color
// when the alpha value is 0.
func MultiplyAlphaBytesPreserveColors(red, green, blue, alpha byte) (r, g, b, a byte) {
	// If alpha is 0 or 0xff, we don't need to multiply. We specifically don't when alpha is 0 because we want to preserve the color components.
	if alpha != 0 && alpha != 0xff {
		// NRGBA.RGBA() returns rOut = rr * a / 0xff, same for green and blue but with gg and bb instead of rr, and returns aOut = aa, where xx means (x | (x << 8)),
		// and r, g, b, a is the input (here red, green, blue, alpha), and rOut etc is the output.
		// It outputs 16-bit components, but we want to output 8-bit components.
		// The 8-bit equivalent would be rOut = r * a / 0xff, same for green and blue, aOut = a
		// That gives different results, however.
		// So in practice it seems that we would have to do rOut = (rr * a / 0xff) >> 8, ..., aOut = aa >> 8.
		r = byte(((uint32(red) | (uint32(red) << 8)) * uint32(alpha) / 0xff) >> 8)
		g = byte(((uint32(green) | (uint32(green) << 8)) * uint32(alpha) / 0xff) >> 8)
		b = byte(((uint32(blue) | (uint32(blue) << 8)) * uint32(alpha) / 0xff) >> 8)
		a = byte((uint32(alpha) | (uint32(alpha) << 8)) >> 8)
	} else {
		r = red
		g = green
		b = blue
		a = alpha
	}
	return
}
