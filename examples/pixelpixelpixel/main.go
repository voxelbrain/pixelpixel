package main

import (
	"image"
	"image/color"
	"math/rand"
	"time"

	"github.com/voxelbrain/pixelpixel/pixelutils"
)

const (
	margin = 256 / 4
)

func main() {
	c := pixelutils.PixelPusher()

	fullPixel := pixelutils.NewPixel()
	subPixel := pixelutils.SubImage(fullPixel, image.Rect(margin, margin, 256-margin, 256-margin))
	square := pixelutils.SubImage(fullPixel, image.Rect(256-margin, 0, 256, margin))
	square = pixelutils.DimensionChanger(square, 3, 5)

	colors := []color.Color{pixelutils.Red, pixelutils.Green, pixelutils.Blue}
	for {
		indices := rand.Perm(len(colors))
		pixelutils.Fill(fullPixel, colors[indices[0]])
		pixelutils.Fill(subPixel, colors[indices[1]])
		pixelutils.DrawText(square, colors[indices[2]], "2")
		c <- fullPixel
		time.Sleep(1 * time.Second)
	}
}
