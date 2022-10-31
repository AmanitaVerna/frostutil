package frostutil

import (
	"errors"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// NewImageFromEImage converts an *ebiten.Image to an *image.RGBA by retrieving the raw RGBA pixel data and copying it to a new image, which it returns.
// We do this so that we can save *ebiten.Images as PNGs, since attempting to directly feed an *ebiten.Image to png.Encode results in garbage output.
func NewImageFromEImage(eImg *ebiten.Image) (img *image.RGBA) {
	left := eImg.Bounds().Min.X
	top := eImg.Bounds().Min.Y
	width := eImg.Bounds().Max.X - left
	height := eImg.Bounds().Max.Y - top
	img = image.NewRGBA(image.Rect(0, 0, width, height))

	// ReadPixels fills pixelBytes with 8bpp image data
	// This retains the alpha data in ebiten 2.4+.
	var pixelBytes []byte = make([]byte, 4*width*height)
	eImg.ReadPixels(pixelBytes)
	CopyImageLines(img.Pix, img.Stride, pixelBytes, width<<2)
	return
}

// NewEImageFromImage converts an image.Image to an *ebiten.Image by creating a new *ebiten.Image and
// writing the image data into it (with the new WritePixels method introduced in ebitengine 2.4.*).
// If mipmaps is true, the *ebiten.Image is created with mipmaps.
// It can relatively quickly handle the source image being an *ebiten.Image, an *image.RGBA, or an *image.NRGBA
// (in which case it converts the NRGBA pixel data to RGBA pixel data, since that is what *ebiten.Images use).
// It can handle other image types, but does it more slowly since it has to copy the image data pixel by pixel.
// I originally wrote this because ebiten.NewImageFromImage was corrupting the pixel data of the source images passed to it
// (I don't know if it still does, but if so, calling this instead should prevent it).
func NewEImageFromImage(img image.Image, mipmaps bool) (ret *ebiten.Image) {
	left := img.Bounds().Min.X
	top := img.Bounds().Min.Y
	width := img.Bounds().Max.X - left
	height := img.Bounds().Max.Y - top
	rect := image.Rect(0, 0, width, height)
	ret = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: !mipmaps})
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
		var col color.NRGBA
		for y := 0; y < height; y++ {
			idx := rowIdx
			for x := 0; x < width; x++ {
				col.R = iImg.Pix[idx]
				col.G = iImg.Pix[idx+1]
				col.B = iImg.Pix[idx+2]
				col.A = iImg.Pix[idx+3]
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
		SlowImageCopy(ret, img)
	}
	return
}

// CopyImage creates a new image with the same width and height as img, and copies its pixel data into it.
// If img is an *ebiten.Image, it creates another *ebiten.Image and uses ReadPixels and WritePixels to copy the image data.
// If it's an *image.RGBA, it creates another *image.RGBA and directly copies the pixel data.
// If it's an *image.NRGBA, it creates another *image.NRGBA and directly copies the pixel data.
// If it's an *image.RGBA or *image.NRGBA and the strides on the input and output images are different (which shouldn't happen unless Go's
// image code changes to make stride something other than width * 4 in RGBA and NRGBA images), then it calls CopyImageLines, which is still reasonably fast.
// If it's any other image type, it creates an *image.RGBA and calls SlowImageCopy, which copies the image data from the input image into the output image pixel by pixel
// using At and Set, which is pretty slow.
// CopyImage returns the copy it creates.
func CopyImage(img image.Image, mipmaps bool) (ret image.Image) {
	left := img.Bounds().Min.X
	top := img.Bounds().Min.Y
	width := img.Bounds().Max.X - left
	height := img.Bounds().Max.Y - top
	rect := image.Rect(0, 0, width, height)
	if eImg, ok := img.(*ebiten.Image); ok {
		var pixelBytes []byte = make([]byte, 4*width*height)
		eImg.ReadPixels(pixelBytes)
		cEImg := ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: !mipmaps})
		cEImg.WritePixels(pixelBytes)
		ret = cEImg
	} else if iImg, ok := img.(*image.RGBA); ok {
		oImg := image.NewRGBA(rect)
		if iImg.Stride == oImg.Stride {
			copy(oImg.Pix, iImg.Pix)
		} else {
			CopyImageLines(oImg.Pix, oImg.Stride, iImg.Pix, iImg.Stride)
		}
		ret = oImg
	} else if iImg, ok := img.(*image.NRGBA); ok {
		oImg := image.NewNRGBA(rect)
		if iImg.Stride == oImg.Stride {
			copy(oImg.Pix, iImg.Pix)
		} else {
			CopyImageLines(oImg.Pix, oImg.Stride, iImg.Pix, iImg.Stride)
		}
		ret = oImg
	} else {
		oImg := image.NewRGBA(rect)
		SlowImageCopy(oImg, iImg)
		ret = oImg
	}
	return
}

// CopyImageLines copies pixel data from iPix to oPix line by line.
// oPix should be the output image's pixel data buffer, oStride should be its Stride,
// and iPix and iStride should be the same for the input image.
func CopyImageLines(oPix []byte, oStride int, iPix []byte, iStride int) {
	oIdx := 0
	lowerStride := Min(oStride, iStride)
	for iIdx := 0; iIdx < len(iPix); iIdx += iStride {
		copy(oPix[oIdx:oIdx+lowerStride], iPix[iIdx:iIdx+lowerStride])
		oIdx += oStride
	}
}

// SlowImageCopy copies pixel data from iImg to oImg pixel by pixel using (Image).At and (Image).Set. It's called by CopyImage or NewEImageFromImage
// if iImg isn't an *ebiten.Image, *image.NRGBA, or *image.RGBA.
// Currently, oImg must still be one of those three for this to work, since the image.Image interface doesn't have a Set method. If it isn't one of those,
// this returns an error.
// For each pixel of each row, it uses the At method to get the pixel color from the source image, and Set to set it on the output image.
// Warning: (*image.NRGBA).Set sets the pixel's color components to 0 when the alpha component is 0,
// even if the color components aren't 0 in the color which was returned by At.
// This causes any tests which attempt to use At and Set to copy pixels with non-zero color components and a zero alpha component to show failures.
// The same is true for *(image.NRGBA64).Set.
func SlowImageCopy(oImg, iImg image.Image) (err error) {
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
	} else {
		err = errors.New("SlowImageCopy only knows how to write to images of type *ebiten.Image, *image.NRGBA, and *image.RGBA. We got a different type instead.")
	}
	return
}
