package main

import (
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"image"
	"time"
)

func main() {
	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()
	grid := pixelutils.DimensionChanger(pixel, 3*4, 3*6)
	for i := 1; i <= 9; i++ {
		pixelutils.FillRectangle(grid, image.Rect(0, 0, 3*4, 3*6), pixelutils.Black)
		pixelutils.DrawText(grid, image.Rect(0, 0, 3*4, 3*6), pixelutils.Red, "123456789"[0:i])
		c <- pixel
		time.Sleep(1 * time.Second)
		if i == 9 {
			i = 0
		}
	}
	select {}
}
