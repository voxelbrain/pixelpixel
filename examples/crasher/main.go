package main

import (
	"fmt"
	"time"

	"github.com/voxelbrain/pixelpixel/pixelutils"
)

func main() {
	wall, _ := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	bigPixel := pixelutils.DimensionChanger(pixel, 16, 12)
	textPixel := pixelutils.NewImageWriter(bigPixel, pixelutils.Red)
	for i := 0; i < 5; i++ {
		if i > 3 {
			panic("CRASH")
		}
		fmt.Fprintf(textPixel, "%d ", 3-i)
		wall <- pixel
		time.Sleep(1 * time.Second)
	}
}
