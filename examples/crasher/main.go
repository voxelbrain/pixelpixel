package main

import (
	"fmt"
	"image"
	"time"

	"github.com/voxelbrain/pixelpixel/pixelutils"
)

func main() {
	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	bigPixel := pixelutils.DimensionChanger(pixel, 4, 6).(pixelutils.Pixel)
	for i := 0; i < 5; i++ {
		color := pixelutils.Green
		if i > 3 {
			panic("CRASH")
		} else if i == 3 {
			color = pixelutils.Red
		}
		pixelutils.Empty(bigPixel)
		pixelutils.DrawText(pixelutils.SubPixel(bigPixel, image.Rect(0, 0, 4, 6)), color, fmt.Sprintf("%d", 3-i))
		c <- pixel
		time.Sleep(1 * time.Second)
	}
}
