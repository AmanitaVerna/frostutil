package frostutil_test

import (
	"errors"
	"fmt"
	"image"
	"testing"

	"github.com/amanitaverna/frostutil"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

const (
	testImgWidth, testImgHeight int = 256, 256
)

type AlphaTestMode int

const (
	Alpha_00                 AlphaTestMode = iota // The entire image is set to alpha 0x00
	Alpha_7F                                      // The entire image is set to alpha 0x7f
	Alpha_FF                                      // The entire image is set to alpha 0xff
	Alpha_HorizontalGradient                      // The image's pixels' alpha values are set to x & 0xff
	Alpha_VerticalGradient                        // The image's pixels' alpha values are set to y & 0xff
	Alpha_DiagonalGradient                        // The image's pixels' alpha values are set to ((x + y) >> 1) & 0xff
	NumAlphaTestModes
)

// GetTestImageNRGBA creates a 256x256 *image.NRGBA where red corresponds to the x pixel coordinate,
// green corresponds to the y pixel coordinate, blue is based on (x+y)>>1 and goes from 0 at the top left to 255 at the bottom-right,
// and alpha is set based on AlphaTestMode.
func GetTestImageNRGBA(alphaTestMode AlphaTestMode) image.Image {
	rect := image.Rect(0, 0, testImgWidth, testImgHeight)
	img := image.NewNRGBA(rect)
	setTestImageNRGBABytes(img.Pix, img.Stride, alphaTestMode)
	return img
}

func setTestImageNRGBABytes(pixelData []byte, stride int, alphaTestMode AlphaTestMode) {
	yIdx := 0
	for y := 0; y < testImgHeight; y++ {
		xIdx := yIdx
		for x := 0; x < testImgWidth; x++ {
			switch alphaTestMode {
			case Alpha_00:
				pixelData[xIdx+3] = 0x00
			case Alpha_7F:
				pixelData[xIdx+3] = 0x7f
			case Alpha_FF:
				pixelData[xIdx+3] = 0xff
			case Alpha_HorizontalGradient:
				pixelData[xIdx+3] = byte(x & 0xff)
			case Alpha_VerticalGradient:
				pixelData[xIdx+3] = byte(y & 0xff)
			case Alpha_DiagonalGradient:
				pixelData[xIdx+3] = byte(((x + y) >> 1) & 0xff)
			}
			pixelData[xIdx] = byte(x & 0xff)
			pixelData[xIdx+1] = byte(y & 0xff)
			pixelData[xIdx+2] = byte(((x + y) >> 1) & 0xff)

			xIdx += 4
		}
		yIdx += stride
	}
}

// GetTestImageRGBA creates a 256x256 *image.RGBA where red corresponds to the x pixel coordinate,
// green corresponds to the y pixel coordinate, blue is based on (x+y)>>1 and goes from 0 at the top left to 255 at the bottom-right,
// and alpha is set based on AlphaTestMode.
func GetTestImageRGBA(alphaTestMode AlphaTestMode) image.Image {
	rect := image.Rect(0, 0, testImgWidth, testImgHeight)
	img := image.NewRGBA(rect)
	setTestImageNRGBABytes(img.Pix, img.Stride, alphaTestMode)
	// multiply red, green, and blue by alpha
	yIdx := 0
	for y := 0; y < testImgHeight; y++ {
		xIdx := yIdx
		for x := 0; x < testImgWidth; x++ {
			img.Pix[xIdx], img.Pix[xIdx+1], img.Pix[xIdx+2], img.Pix[xIdx+3] = frostutil.MultiplyAlphaBytes(img.Pix[xIdx], img.Pix[xIdx+1], img.Pix[xIdx+2], img.Pix[xIdx+3])
			xIdx += 4
		}
		yIdx += img.Stride
	}
	return img
}

// CheckImagePattern returns nil if the image matches the pattern described in the comment for GetTestImage.
// If it doesn't match, it returns an error indicating what's wrong.
// This is only implemented for *ebiten.Image, *image.RGBA, and *image.NRGBA.
func CheckImagePattern(img image.Image, alphaTestMode AlphaTestMode) (err error) {
	if img == nil {
		return errors.New("CheckImagePattern was passed a nil image!")
	}
	// first, verify bounds match
	width, height := testImgWidth, testImgHeight
	rect := image.Rect(0, 0, width, height)
	if img.Bounds() != rect {
		err = errors.New(fmt.Sprintf("Image bounds do not match. Expected %v, got %v.", rect, img.Bounds()))
	} else {
		var pixelData []byte
		if eImg, ok := img.(*ebiten.Image); ok {
			pixelData = make([]byte, width*height*4)
			eImg.ReadPixels(pixelData)
			err = checkImagePatternImpl(pixelData, width*4, width*4, alphaTestMode, true)
		} else if xImg, ok := img.(*image.RGBA); ok {
			pixelData = xImg.Pix
			err = checkImagePatternImpl(pixelData, width*4, xImg.Stride, alphaTestMode, true)
		} else if xImg, ok := img.(*image.NRGBA); ok {
			pixelData = xImg.Pix
			err = checkImagePatternImpl(pixelData, width*4, xImg.Stride, alphaTestMode, false)
		}
	}
	return
}

// checkImagePatternImpl implements the logic of CheckImagePattern.
func checkImagePatternImpl(pixelData []byte, rowBytes, stride int, alphaTestMode AlphaTestMode, isRGBA bool) (err error) {
	y := 0
	var red, green, blue, alpha byte
	for yIdx := 0; yIdx < len(pixelData); yIdx += stride {
		x := 0
		for idx := yIdx; idx < yIdx+rowBytes; idx += 4 {
			red = byte(x & 0xff)
			green = byte(y & 0xff)
			blue = byte(((x + y) >> 1) & 0xff)
			switch alphaTestMode {
			case Alpha_00:
				alpha = 0x00
			case Alpha_7F:
				alpha = 0x7f
			case Alpha_FF:
				alpha = 0xff
			case Alpha_HorizontalGradient:
				alpha = byte(x & 0xff)
			case Alpha_VerticalGradient:
				alpha = byte(y & 0xff)
			case Alpha_DiagonalGradient:
				alpha = byte(((x + y) >> 1) & 0xff)
			}
			if isRGBA {
				red, green, blue, alpha = frostutil.MultiplyAlphaBytes(red, green, blue, alpha)
			}
			if pixelData[idx] != red || pixelData[idx+1] != green || pixelData[idx+2] != blue || pixelData[idx+3] != alpha {
				err = errors.New(fmt.Sprintf("Pixel [%v, %v]: Expected color #%02x%02x%02x%02x, got #%02x%02x%02x%02x.", x, y, red, green, blue, alpha, pixelData[idx], pixelData[idx+1], pixelData[idx+2], pixelData[idx+3]))
			}
			x += 1
		}
		y += 1
	}
	return
}

// Tests NewEImageFromImage and NewImageFromEImage
func Test_ImageConversion(t *testing.T) {
	frostutil.QueueUpdateTest(t, test_ImageConversion)
}

func test_ImageConversion(t *testing.T) {
	ass := assert.New(t)
	for alphaTestMode := AlphaTestMode(0); alphaTestMode < NumAlphaTestModes; alphaTestMode++ {
		img := GetTestImageNRGBA(alphaTestMode)
		test_ImageConversionImpl(ass, img, alphaTestMode)

		img = GetTestImageRGBA(alphaTestMode)
		test_ImageConversionImpl(ass, img, alphaTestMode)
	}
}

func test_ImageConversionImpl(ass *assert.Assertions, img image.Image, alphaTestMode AlphaTestMode) {
	ass.NotNil(img)
	imgType := "[unknown image type]"
	if _, ok := img.(*image.NRGBA); ok {
		imgType = "NRGBA"
	} else if _, ok := img.(*image.RGBA); ok {
		imgType = "RGBA"
	} else if _, ok := img.(*ebiten.Image); ok {
		imgType = "*ebiten.Image"
	}
	var err error
	if err = CheckImagePattern(img, alphaTestMode); err != nil {
		ass.Fail(fmt.Sprintf("Image returned by GetTestImage%v failed to match expected pattern", imgType), err.Error())
	} else {
		ass.Nil(err)
	}
	eImg := frostutil.NewEImageFromImage(img, false)
	ass.NotNil(eImg)
	if err = CheckImagePattern(eImg, alphaTestMode); err != nil {
		ass.Fail(fmt.Sprintf("NewEImageFromImage returned an *ebiten.Image which failed to match expected pattern with alphaTestMode=%v and image type %v", alphaTestMode, imgType), err.Error())
	} else {
		ass.Nil(err)
	}
	img2 := frostutil.NewImageFromEImage(eImg)
	ass.NotNil(img2)
	if err = CheckImagePattern(img2, alphaTestMode); err != nil {
		ass.Fail(fmt.Sprintf("NewImageFromEImage returned an *image.NRGBA which failed to match expected pattern with alphaTestMode=%v and imgType %v", alphaTestMode, imgType), err.Error())
	} else {
		ass.Nil(err)
	}

}

// Tests CopyImage. We want to verify that it correctly copies *ebiten.Image, *image.NRGBA, and *image.RGBA images.
func Test_CopyImage(t *testing.T) {
	frostutil.QueueUpdateTest(t, test_CopyImage)
}

// Tests CopyImage. We want to verify that it correctly copies *ebiten.Image, *image.NRGBA, and *image.RGBA images.
func test_CopyImage(t *testing.T) {
	// *image.NRGBA:
	ass := assert.New(t)
	for alphaTestMode := AlphaTestMode(0); alphaTestMode < NumAlphaTestModes; alphaTestMode++ {
		nImg := GetTestImageNRGBA(alphaTestMode)
		test_CopyImageImpl(ass, nImg, alphaTestMode)

		// *image.RGBA:
		rImg := GetTestImageRGBA(alphaTestMode)
		test_CopyImageImpl(ass, rImg, alphaTestMode)

		// *ebiten.Image:
		// convert the RGBA image to *ebiten.Image to get an *ebiten.Image to copy
		eImg := frostutil.NewEImageFromImage(rImg, false)
		test_CopyImageImpl(ass, eImg, alphaTestMode)
	}
}

// Since we do the exact same thing with all three image types, this function contains the duplicated code.
func test_CopyImageImpl(ass *assert.Assertions, img image.Image, alphaTestMode AlphaTestMode) {
	ass.NotNil(img)
	imgType := "[unknown image type]"
	if _, ok := img.(*image.NRGBA); ok {
		imgType = "NRGBA"
	} else if _, ok := img.(*image.RGBA); ok {
		imgType = "RGBA"
	} else if _, ok := img.(*ebiten.Image); ok {
		imgType = "*ebiten.Image"
	}
	var err error
	if err = CheckImagePattern(img, alphaTestMode); err != nil {
		ass.Fail(fmt.Sprintf("Image returned by GetTestImage%v failed to match expected pattern", imgType), err.Error())
	} else {
		ass.Nil(err)
	}
	cImg := frostutil.CopyImage(img, false)
	if err = CheckImagePattern(cImg, alphaTestMode); err != nil {
		ass.Fail(fmt.Sprintf("CopyImage failed to correctly copy our %v test image", imgType), err.Error())
	} else {
		ass.Nil(err)
	}
}

// Test CopyImageLines
func Test_CopyImageLines(t *testing.T) {
	var err error
	ass := assert.New(t)
	for alphaTestMode := AlphaTestMode(0); alphaTestMode < NumAlphaTestModes; alphaTestMode++ {
		img := GetTestImageNRGBA(alphaTestMode).(*image.NRGBA)
		ass.NotNil(img)
		if err = CheckImagePattern(img, alphaTestMode); err != nil {
			ass.Fail("Image returned by GetTestImageNRGBA failed to match expected pattern", err.Error())
		} else {
			ass.Nil(err)
		}
		cImg := image.NewNRGBA(img.Bounds())
		ass.NotNil(cImg)
		frostutil.CopyImageLines(cImg.Pix, cImg.Stride, img.Pix, img.Stride)
		if err = CheckImagePattern(cImg, alphaTestMode); err != nil {
			ass.Fail(fmt.Sprintf("CopyImageLines failed to correctly copy our test image's pixel data bytes"), err.Error())
		} else {
			ass.Nil(err)
		}
	}
}

// Test SlowImageCopy
func Test_SlowImageCopy(t *testing.T) {
	frostutil.QueueUpdateTest(t, test_SlowImageCopy)
}

// Tests SlowImageCopy. We want to verify that it correctly copies *ebiten.Image, *image.NRGBA, and *image.RGBA images.
func test_SlowImageCopy(t *testing.T) {
	// *image.NRGBA:
	ass := assert.New(t)
	for alphaTestMode := AlphaTestMode(0); alphaTestMode < NumAlphaTestModes; alphaTestMode++ {
		nImg := GetTestImageNRGBA(alphaTestMode)
		test_SlowImageCopyImpl(ass, nImg, alphaTestMode)

		// *image.RGBA:
		rImg := GetTestImageRGBA(alphaTestMode)
		test_SlowImageCopyImpl(ass, rImg, alphaTestMode)

		// *ebiten.Image:
		// convert the RGBA image to *ebiten.Image to get an *ebiten.Image to copy
		eImg := frostutil.NewEImageFromImage(rImg, false)
		test_SlowImageCopyImpl(ass, eImg, alphaTestMode)
	}
}

// Since we do the exact same thing with all three image types, this function contains the duplicated code.
func test_SlowImageCopyImpl(ass *assert.Assertions, img image.Image, alphaTestMode AlphaTestMode) {
	ass.NotNil(img)
	imgType := "[unknown image type]"
	if _, ok := img.(*image.NRGBA); ok {
		imgType = "NRGBA"
	} else if _, ok := img.(*image.RGBA); ok {
		imgType = "RGBA"
	} else if _, ok := img.(*ebiten.Image); ok {
		imgType = "*ebiten.Image"
	}
	var err error
	if err = CheckImagePattern(img, alphaTestMode); err != nil {
		ass.Fail(fmt.Sprintf("Image returned by GetTestImage%v failed to match expected pattern", imgType), err.Error())
	} else {
		ass.Nil(err)
	}
	cImg := image.NewRGBA(img.Bounds())
	ass.NotNil(cImg)
	frostutil.SlowImageCopy(cImg, img)
	if err = CheckImagePattern(cImg, alphaTestMode); err != nil {
		ass.Fail(fmt.Sprintf("SlowImageCopy failed to correctly copy our %v test image with alphaTestMode=%v ", imgType, alphaTestMode), err.Error())
	} else {
		ass.Nil(err)
	}
}
