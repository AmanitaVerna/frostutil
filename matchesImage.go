package frostutil

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expectedFolder string = "testdata/expected"
	failedFolder   string = "testdata/failed"
	pngStr         string = ".png"
)

// MatchesImage compares an image.Image to "testdata/expected/<imageName>.png". If img is not nil, it attempts to open "testdata/expected/<imageName>.png".
// If it succeeds, it converts it to an image.Image, and then compares the two images.
// If it fails, it writes the image to "testdata/failed/<imageName>.png" and raises a test failure.
// It can handle *ebiten.Images and save them as PNGs.
// Also returns true if the images match, and false if they don't.
func MatchesImage(t *testing.T, imageName string, img image.Image) bool {
	if assert.NotNil(t, img) {
		filename := expectedFolder + "/" + imageName + pngStr
		fr, err := os.Open(filename)
		failedBuilder := &strings.Builder{}
		if err != nil {
			if os.IsNotExist(err) {
				failedBuilder.WriteString(filename)
				failedBuilder.WriteString(" doesn't exist.")
			} else {
				failedBuilder.WriteString(fmt.Sprintf("os.Open(%v) failed: %v", filename, err))
			}
		} else {
			require.NotNil(t, fr)
			defer fr.Close()
			r := bufio.NewReader(fr)
			pngImg, err := png.Decode(r)
			require.NoError(t, err)
			require.NotNil(t, pngImg)
			bounds := img.Bounds()
			if pngImg.Bounds().Dx() != bounds.Dx() || pngImg.Bounds().Dy() != bounds.Dy() {
				failedBuilder.WriteString(fmt.Sprintf("Dimensions of %v (%v, %v) don't match. Expected (%v, %v).\n", imageName, bounds.Dx(), bounds.Dy(), pngImg.Bounds().Dx(), pngImg.Bounds().Dy()))
			} else {
				for y := 0; y < bounds.Dy() && failedBuilder.Len() < 1000; y++ {
					for x := 0; x < bounds.Dx() && failedBuilder.Len() < 1000; x++ {
						c1 := img.At(x+bounds.Min.X, y+bounds.Min.Y)
						c2 := pngImg.At(x+pngImg.Bounds().Min.X, y+pngImg.Bounds().Min.Y)
						r1, g1, b1, a1 := ToNRGBA(c1)
						r2, g2, b2, a2 := ToNRGBA(c2)
						if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
							r1a, g1a, b1a, a1a := c1.RGBA()
							failedBuilder.WriteString("Pixel (")
							failedBuilder.WriteString(fmt.Sprintf("%v, %v", x, y))
							failedBuilder.WriteString(") of ")
							failedBuilder.WriteString(imageName)
							failedBuilder.WriteString(" doesn't match. Got NRGBA #")
							failedBuilder.WriteString(fmt.Sprintf("%02x%02x%02x%02x, expected NRGBA #%02x%02x%02x%02x. *ebiten.Image's RGBA color here is %04x%04x%04x%04x\n", r1, g1, b1, a1, r2, g2, b2, a2, r1a, g1a, b1a, a1a))
						}
					}
				}
			}
			if failedBuilder.Len() > 1000 {
				failedBuilder.WriteString("...")
			}
		}
		failed := failedBuilder.String()
		if len(failed) > 0 {
			failedFilename := failedFolder + "/" + imageName + pngStr
			os.MkdirAll(failedFolder, 0644) //read and write permissions for the owner, read-only for group and others
			fw, err := os.Create(failedFilename)
			require.NoError(t, err)
			defer fw.Close()
			w := bufio.NewWriter(fw)
			eImg, isEImg := img.(*ebiten.Image)
			if isEImg {
				png.Encode(w, NewImageFromEImage(eImg))
			} else {
				png.Encode(w, img)
			}
			w.Flush()
			assert.Fail(t, failed)
		}
		return len(failed) == 0
	} else {
		return false
	}
}
