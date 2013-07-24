package imageutils

import (
	"image"
	"image/color"
	"image/draw"
)

func FillRectangle(img draw.Image, r image.Rectangle, c color.Color) {
	fillColor := &image.Uniform{c}
	draw.Draw(img, r, fillColor, image.Point{0, 0}, draw.Over)
}
