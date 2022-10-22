package frostutil

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// NewImageFromEImage converts an *ebiten.Image to an *image.NRGBA by retrieving the raw RGBA pixel data, unmultiplying the alpha, and storing the resulting values
// in img.
// It's slow. We use it so that we can save *ebiten.Images as PNGs.
func NewImageFromEImage(eImg *ebiten.Image) (img *image.NRGBA) {
	left := eImg.Bounds().Min.X
	top := eImg.Bounds().Min.Y
	width := eImg.Bounds().Max.X - left
	height := eImg.Bounds().Max.Y - top
	img = image.NewNRGBA(image.Rect(0, 0, width, height))

	// ReadPixels fills pixelBytes with 8bpp image data
	// This retains the alpha data in ebiten 2.4+.
	var pixelBytes []byte = make([]byte, 4*width*height)
	eImg.ReadPixels(pixelBytes)
	inIdxRow := 0
	outIdxRow := 0
	stride := img.Stride
	for y := 0; y < height; y++ {
		outIdx := outIdxRow
		inIdx := inIdxRow
		for x := 0; x < width; x++ {
			img.Pix[outIdx], img.Pix[outIdx+1], img.Pix[outIdx+2], img.Pix[outIdx+3] = UnmultiplyAlphaBytes(pixelBytes[inIdx], pixelBytes[inIdx+1], pixelBytes[inIdx+2], pixelBytes[inIdx+3])
			outIdx += 4
			inIdx += 4
		}
		outIdxRow += stride
		inIdxRow += width << 2
	}

	/*
		// Somehow this loses the alpha byte in ebitengine 2.4 - (at least) 2.4.7 on directx.
		// This happens even if we don't run at through ToNRGBA_Color (that was an attempt to fix the problem).
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				at := eImg.At(left+x, top+y)
				img.Set(x, y, ToNRGBA_Color(at))
			}
		}
	*/
	return
}

// NewEImageFromImage converts an image.Image to an *ebiten.Image by creating a new *ebiten.Image and
// copying the image data into it, rather than by calling ebiten.NewImageFromImage.
// We do this to prevent an image corruption issue that we've been experiencing.
func NewEImageFromImage(img image.Image) (ret *ebiten.Image) {
	left := img.Bounds().Min.X
	top := img.Bounds().Min.Y
	width := img.Bounds().Max.X - left
	height := img.Bounds().Max.Y - top
	rect := image.Rect(0, 0, width, height)
	ret = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})
	// copy the image data
	if eImg, ok := img.(*ebiten.Image); ok {
		var pixelBytes []byte = make([]byte, 4*width*height)
		eImg.ReadPixels(pixelBytes)
		ret.WritePixels(pixelBytes)
	} else if iImg, ok := img.(*image.RGBA); ok {
		ret.WritePixels(iImg.Pix)
	} else if iImg, ok := img.(*image.NRGBA); ok {
		// we need to convert the pixel data to RGBA
		pixelBytes := make([]byte, width*height*4)
		rowIdx := 0
		for y := 0; y < height; y++ {
			idx := rowIdx
			for x := 0; x < width; x++ {
				col := color.NRGBA{R: iImg.Pix[idx], G: iImg.Pix[idx+1], B: iImg.Pix[idx+2], A: iImg.Pix[idx+3]}
				r, g, b, a := col.RGBA() // get alpha-premultiplied rgba values
				pixelBytes[idx] = byte(r >> 8)
				pixelBytes[idx+1] = byte(g >> 8)
				pixelBytes[idx+2] = byte(b >> 8)
				pixelBytes[idx+3] = byte(a >> 8)
				idx += 4
			}
			rowIdx += iImg.Stride
		}
		ret.WritePixels(pixelBytes)
	} else {
		slowCopy(ret, img)
	}
	return
	// ebitengine 2.3.8 and before:
	// return ebiten.NewImageFromImage(img)
	// ebitengine 2.4.0 and later:
	// return ebiten.NewImageFromImageWithOptions(img, &ebiten.NewImageFromImageOptions{Unmanaged: true})
}

// CopyImage creates a new image with the same width and height as img, and copies its pixel data into it.
// If img is an *ebiten.Image, it creates another *ebiten.Image and uses ReadPixels and WritePixels.
// If it's an *image.RGBA, it creates another *image.RGBA and directly copies the pixel data.
// If it's an *image.NRGBA, it creates another *image.NRGBA and directly copies the pixel data.
// If it's anything else, or it's an *image.RGBA or *image.NRGBA and the strides on the input and output images are different (which shouldn't happen unless Go's
// image code changes to make stride something other than width * 4 in RGBA and NRGBA images), then it calls slowCopy, which
// copies the image data pixel by pixel using At and Set.
// It returns the copy.
func CopyImage(img image.Image) (ret image.Image) {
	left := img.Bounds().Min.X
	top := img.Bounds().Min.Y
	width := img.Bounds().Max.X - left
	height := img.Bounds().Max.Y - top
	rect := image.Rect(0, 0, width, height)
	if eImg, ok := img.(*ebiten.Image); ok {
		var pixelBytes []byte = make([]byte, 4*width*height)
		eImg.ReadPixels(pixelBytes)
		cEImg := ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})
		cEImg.WritePixels(pixelBytes)
		ret = cEImg
	} else if iImg, ok := img.(*image.RGBA); ok {
		oImg := image.NewRGBA(rect)
		if iImg.Stride == oImg.Stride {
			copy(oImg.Pix, iImg.Pix)
		} else {
			slowCopy(oImg, iImg)
		}
		ret = oImg
	} else if iImg, ok := img.(*image.NRGBA); ok {
		oImg := image.NewNRGBA(rect)
		if iImg.Stride == oImg.Stride {
			copy(oImg.Pix, iImg.Pix)
		} else {
			slowCopy(oImg, iImg)
		}
		ret = oImg
	} else {
		oImg := image.NewNRGBA(rect)
		slowCopy(oImg, iImg)
		ret = oImg
	}
	return
}

// slowCopy copies pixel data from iPix to oPix row by row. It's called by CopyImage if the images have different strides,
// and by NewEImageFromImage if iImg isn't an *ebiten.Image, *image.NRGBA, or *image.RGBA.
func slowCopy(oImg, iImg image.Image) {
	left := iImg.Bounds().Min.X
	top := iImg.Bounds().Min.Y
	width := oImg.Bounds().Dx()
	height := oImg.Bounds().Dy()
	if xOImg, ok := oImg.(*image.RGBA); ok {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				xOImg.Set(x, y, iImg.At(x+left, y+top))
			}
		}
	} else if xOImg, ok := oImg.(*image.NRGBA); ok {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				xOImg.Set(x, y, iImg.At(x+left, y+top))
			}
		}
	} else if xOImg, ok := oImg.(*ebiten.Image); ok {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				xOImg.Set(x, y, iImg.At(x+left, y+top))
			}
		}
	}
}
