package pixelutils

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
)

// ImageWriter is a writable image. The image behaves like a
// non-scrolling terminal.
type ImageWriter interface {
	draw.Image
	io.Writer
	// Resets the `cursor` to the initial position
	Cls()
}

type imageWriter struct {
	draw.Image
	charRect image.Rectangle
	color    color.Color
}

// NewImageWriter creates a new ImageWriter, which writes on the
// given image in the given color.
func NewImageWriter(img draw.Image, c color.Color) ImageWriter {
	iw := &imageWriter{
		Image: img,
		color: c,
	}
	iw.Cls()
	return iw
}

func (iw *imageWriter) Write(text []byte) (int, error) {
	r := iw.Bounds().Canon()
	if !r.Overlaps(iw.charRect) {
		return 0, fmt.Errorf("Write region is not inside image anymore")
	}
	colorImg := &image.Uniform{iw.color}
	for _, char := range []byte(text) {
		if char == '\n' {
			nextLine(&iw.charRect, r.Min.X)
			continue
		}
		mask := MaskForCharacter(char)
		draw.DrawMask(iw, iw.charRect, colorImg, image.Point{0, 0}, mask, image.Point{0, 0}, draw.Over)
		iw.charRect = iw.charRect.Add(image.Point{4, 0})
		if !iw.charRect.In(r) {
			nextLine(&iw.charRect, r.Min.X)
		}
	}
	return len(text), nil
}

func (iw *imageWriter) Cls() {
	r := iw.Image.Bounds().Canon()
	iw.charRect = image.Rectangle{
		Min: r.Min,
		Max: r.Min.Add(image.Point{4, 6}),
	}
}

// DrawText is a convenience function to write text into an image.
func DrawText(img draw.Image, c color.Color, text string) {
	io.WriteString(NewImageWriter(img, c), text)
}

func nextLine(r *image.Rectangle, xstart int) {
	r.Min.X = xstart
	r.Min.Y += 6
	r.Max = r.Min.Add(image.Point{4, 6})
}

// MaskForCharacter returns an 4x6 black-and-white image mask for
// the given character. Useful for image/draw.DrawMask().
func MaskForCharacter(c byte) image.Image {
	mask := image.NewAlpha(image.Rect(0, 0, 4, 6))
	rows := font[int(c)]
	for y, row := range rows {
		for x := uint(0); x < 4; x++ {
			if (row>>x)&1 > 0 {
				mask.Set(4-(int(x)+1), y, color.Black)
			}
		}
	}
	return mask
}
