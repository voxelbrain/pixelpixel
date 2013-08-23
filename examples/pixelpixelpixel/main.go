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
	wall, clicks := pixelutils.PixelPusher()

	fullPixel := pixelutils.NewPixel()
	subPixel := pixelutils.SubImage(fullPixel, image.Rect(margin, margin, 256-margin, 256-margin))
	square := pixelutils.SubImage(fullPixel, image.Rect(256-margin, 0, 256, margin))
	textSquare := pixelutils.DimensionChanger(square, 3, 5)
	colors := []color.Color{pixelutils.Red, pixelutils.Green, pixelutils.Blue}
	offset := []int{0, 1, 2}

	drawSignal := make(chan bool)

	go func() {
		for i := 0; i < 5; i++ {
			offset = rand.Perm(3)
			drawSignal <- true
			time.Sleep(200 * time.Millisecond)
		}
		for click := range clicks {
			if click.Point().In(subPixel.Bounds()) {
				offset[1] = (offset[1] + 1) % len(colors)
			} else if click.Point().In(square.Bounds()) {
				offset[2] = (offset[2] + 1) % len(colors)
			} else {
				offset[0] = (offset[0] + 1) % len(colors)
			}
			drawSignal <- true
		}
	}()

	for _ = range drawSignal {
		pixelutils.Fill(fullPixel, colors[offset[0]])
		pixelutils.Fill(subPixel, colors[offset[1]])
		pixelutils.DrawText(textSquare, colors[offset[2]], "2")
		wall <- fullPixel
	}
}
