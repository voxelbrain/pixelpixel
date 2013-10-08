package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

var (
	blackPixel = &bytes.Buffer{}
)

func init() {
	r := image.Rect(0, 0, 256, 256)
	blackImage := image.NewRGBA(r)
	draw.Draw(blackImage, r, &image.Uniform{color.Black}, image.Point{0, 0}, draw.Over)

	png.Encode(blackPixel, blackImage)
}
