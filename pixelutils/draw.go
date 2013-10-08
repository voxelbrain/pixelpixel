package pixelutils

import (
	"image"
	"image/color"
	"image/draw"
)

// Fill fills an entire image with the given color.
func Fill(img draw.Image, c color.Color) {
	fillColor := &image.Uniform{c}
	draw.Draw(img, img.Bounds(), fillColor, image.Point{0, 0}, draw.Over)
}

// Empty makes an image black.
func Empty(img draw.Image) {
	Fill(img, Black)
}
