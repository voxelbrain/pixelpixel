package main

import (
	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
	"image"
	"time"
)

func main() {
	c := protocol.PixelPusher()
	pixel := protocol.NewPixel()
	grid := imageutils.DimensionChanger(pixel, 3*4, 3*6)
	for i := 1; i <= 9; i++ {
		imageutils.FillRectangle(grid, image.Rect(0, 0, 3*4, 3*6), imageutils.Black)
		imageutils.DrawText(grid, image.Rect(0, 0, 3*4, 3*6), imageutils.Red, "123456789"[0:i])
		c <- pixel
		time.Sleep(1 * time.Second)
		if i == 9 {
			i = 0
		}
	}
	select {}
}
