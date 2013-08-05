package pixelutils

import (
	"image"
	"image/color"
	"image/draw"
)

func Fill(img draw.Image, c color.Color) {
	fillColor := &image.Uniform{c}
	draw.Draw(img, img.Bounds(), fillColor, image.Point{0, 0}, draw.Over)
}

func Empty(img draw.Image) {
	Fill(img, Black)
}
