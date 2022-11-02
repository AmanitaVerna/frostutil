This package contains a number of utility functions related to math, strings, functional things (map, compose, partial application), colors, and images (*ebiten.Image, *image.NRGBA, and *image.RGBA), as well as a framework to enable running tests under Ebitengine:

In util.go:
- Min and Max functions for all signed or unsigned integers and floats, using generics.
- Abs function for signed integers or floats, using generics.
- Split function that splits a string by a separator rune, but not when that rune is prefixed by a backslash. It also unescapes escaped separators and endlines.
- Join function that does the opposite of Split: It takes a slice of strings, and joins them with a specified separator, while escaping all already-existing instances of that separator and of endlines.
- UnescapeStr and EscapeStr, which just do the separator and endline escaping/unescaping without the splitting/joining.
- SearchStringsUnsorted, which searches an unsorted slice of strings to see if any of them are identical to a particular string, returning the index if it is found. This doesn't sort the slice and I don't recommend using it often, because it is nowhere near as efficient as using a proper search algorithm to search a sorted slice of strings.
- Xor for two boolean values, because Go lacks a boolean xor operator.
- DegreesToRadians and RadiansToDegrees convenience functions.
- A Compose function which composes two functions which each take a single parameter and return value of the same type. That is, it takes (f (T) T) and (g (T) T) and returns a function which takes a parameter x of type T, which when called will call f(g(x)) and return whatever it returns.
- A Map function which takes a function parameter f and a slice parameter xs, and returns a new slice ys created from the results of calling the function f on each element x of the slice xs.
- A Partial2to1 function which does partial application of a two-parameter function, which could be useful for use with Compose or Map. I would like to have written a function that could do partial application of any arbitary function, but I'm pretty sure that's impossible. It would require being able to specify a function with an arbitrary number of parameters (which doesn't use ...) and return values. With this example, though, it should be easy to write a function to do partial application for functions with any arbitary number of parameters etc.

In color.go, a number of functions for 32bpp colors (8 bit color and alpha components):
- ToNRGBA, which converts any arbitary standard color.Color type to NRGBA and returns the individual red, green, blue, and alpha bytes, without using the model convert functions. Unlike the color package's conversion stuff, when alpha is zero, this still preserves the color components, rather than returning them as zero also. Since normally the color components are zeroed when using the model conversion functions or NRGBA color's RGBA() method when alpha is zero, this capability is only relevant when you're manually modifying an image's pixel buffer. If you want to preserve color information when encoding from NRGBA to RGBA, you can use MultiplyAlphaBytesPreserveColors.
- ToNRGBA_Color, which is the same but returns a color.Color.
- ToNRGBA_U32, which is the same but packages the output into a uint32 where the bytes are from high to low: red, green, blue, and alpha.
- UnmultiplyAlpha, which is what ToNRGBA calls when the color it is given is RGBA or RGBA16 or an unrecognized Color type. It can be called directly to skip the switch statement in ToNRGBA. This also preserves color components when alpha is zero. It calls the color's RGBA() method to get the color bytes in RGBA format, then unmultiplies the alpha, if it is neither zero nor 0xffff (RGBA() returns 16-bit color and alpha components).
- UnmultiplyAlphaBytes, which takes the four components as alpha-premultiplied bytes, rather than a color.Color, and returns four components as bytes with the alpha premultiplication removed. That is to say, it converts RGBA bytes to NRGBA bytes. This also preserves color components when alpha is zero. Unlike UnmultiplyAlpha, it doesn't work with arbitary color types.
- MultiplyAlphaBytes, which does the opposite of UnmultiplyAlphaBytes: It converts NRGBA bytes to RGBA bytes. This is more of a convenience function so you don't have to write code to pack the bytes into a color, call RGBA(), and then unpack them.
- MultiplyAlphaBytesPreserveColors, which is like MultiplyAlphaBytes but it preserves the color components when the alpha component is zero. It does the math itself rather than calling RGBA(). It gives results that match what you get from MultiplyAlphaBytes() except for when alpha is 0.

In image.go:
- NewImageFromEImage, which converts an *ebiten.Image to an *image.RGBA by retrieving the raw RGBA pixel data and copying it to a new image, which it returns. We do this so that we can save *ebiten.Images as PNGs, since attempting to directly feed an *ebiten.Image to png.Encode results in garbage output. This is useful for screenshots, for example.
- NewEImageFromImage, which creates a new *ebiten.Image from an arbitary image type. It does so by creating a new *ebiten.Image of the same size, and copying the pixel data from the source image. In theory, that's what ebiten.NewImageFromImage should also be doing, but in practice I found that it was for some reason corrupting the source images fed to it (in Ebitengine 2.3.\*, anyways). This has a bool mipmaps parameter, so that you can easily say whether the new image should have mipmaps or not. Also, this is designed to be able to quickly and efficiently copy *ebiten.Images and *image.RGBA images. It copies *image.NRGBA images more slowly, since it has to convert their pixel data to RGBA before it can copy it to the new *ebiten.Image. Any other image type is copied very slowly, since it won't have access to the pixel data buffer, and will have to copy pixels one by one using At and Set.
- CopyImage, which quickly and efficiently copies an image's pixel data to a new image of the same type (*ebiten.Image, *image.NRGBA, or *image.RGBA) and returns the copy. If given any other type of image, it creates a new *image.RGBA and copies the pixel data into it very slowly using At and Set.
- CopyImageLines copies image data line by line. It is slower than copying the entire pixel data buffer at once, but useful if the source and destination images have different strides (because of padding, for instance). As far as I know, this shouldn't come up with images loaded from PNGs, but it might with other image formats.
- SlowImageCopy copies pixel data from iImg to oImg pixel by pixel using (Image).At and (Image).Set. It's called by CopyImage or NewEImageFromImage if iImg isn't an *ebiten.Image, *image.NRGBA, or *image.RGBA. Since image.Image doesn't have a Set method, oImg must still be one of those three for this to work. If it isn't one of those, it returns an error.

In matchesImage.go:
- MatchesImage, which takes a *testing.T, an image name, and an image, and compares the image to the expected output (which should be a .png file in testdata/expected). If it fails to match, or the expected image is missing, it reports a failure to the *testing.T, attempts to write the failed image to testdata/failed (creating the folder if it doesn't exist), and returns false. If it matches, it returns true. It accepts both regular images and *ebiten.Images. If you just didn't have an expected image yet and it is correct, you can move the output image from testdata/failed to testdata/expected and the next run should pass, assuming the output is the same every time.

Finally, test.go contains the code that enables testing things under Ebitengine in the Layout, Update, and Draw methods. To use this, every package that needs to test things under Ebitengine first needs a single file whose name should start with "test" which contains this function:
```go
func TestMain(m *testing.M) {
	frostutil.OnTestMain(m)
}
```
I personally keep this in main_test.go. You can see one in this package, since image_test.go includes tests that run under Ebitengine.

OnTestMain ensures that every test run in the package under Ebitengine runs in the main/OS thread. It sets up and runs Ebitengine, runs your test functions (via m.Run) (which should call QueueLayoutTest, QueueUpdateTest, and/or QueueDrawTest if they need to run test code under Layout, Update, or Draw), waits for m.Run and all the tests to finish, and then prompts Update to tell Ebitengine to shut down.

And then for the test files you have where you want to test things under Ebitengine, you can write tests like so (note that you don't have to queue them all from one test function, this is just to show all three Queue functions):
```go
func Test_SomeTests(t *testing.T) {
	frostutil.QueueLayoutTest(t, test_SomeLayoutTest)
	frostutil.QueueUpdateTest(t, test_SomeUpdateTest)
	frostutil.QueueDrawTest(t, test_SomeDrawTest)	
}

func test_SomeLayoutTest(t *testing.T, outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// Your actual test here. Also return screen dimensions for the Layout function to return:
	return outsideWidth, outsideHeight // you can return something else if you like
}

func test_SomeUpdateTest(t *testing.T) {
	// Your actual test here
}

func test_SomeDrawTest(t *testing.T, screen *ebiten.Image) {
	// Your actual test here
}
```

You can have multiple tests per file, of course, and they can each queue as many things as they want.
