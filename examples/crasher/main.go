package main

import (
	"fmt"
	"image"
	"time"

	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
)

func main() {
	c := protocol.PixelPusher()
	img := protocol.NewPixel()

	dImg := imageutils.DimensionChanger(img, 4, 6)
	for i := 0; i < 5; i++ {
		color := imageutils.Green
		if i > 3 {
			panic("CRASH")
		} else if i == 3 {
			color = imageutils.Red
		}
		imageutils.FillRectangle(dImg, image.Rect(0, 0, 4, 6), imageutils.Black)
		imageutils.DrawText(dImg, image.Rect(0, 0, 4, 6), color, fmt.Sprintf("%d", 3-i))
		c <- img
		time.Sleep(1 * time.Second)
	}
}
