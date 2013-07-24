package main

import (
	"fmt"
	"image"
	"time"

	"github.com/voxelbrain/pixelpixel/pixelutils"
)

func main() {
	c := pixelutils.PixelPusher()
	img := pixelutils.NewPixel()

	dImg := pixelutils.DimensionChanger(img, 4, 6)
	for i := 0; i < 5; i++ {
		color := pixelutils.Green
		if i > 3 {
			panic("CRASH")
		} else if i == 3 {
			color = pixelutils.Red
		}
		pixelutils.FillRectangle(dImg, image.Rect(0, 0, 4, 6), pixelutils.Black)
		pixelutils.DrawText(dImg, image.Rect(0, 0, 4, 6), color, fmt.Sprintf("%d", 3-i))
		c <- img
		time.Sleep(1 * time.Second)
	}
}
