package main

import (
	"fmt"
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"time"
)

func main() {
	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()
	bigPixel := pixelutils.DimensionChanger(pixel, 5*4, 18)
	textPixel := pixelutils.NewImageWriter(bigPixel, pixelutils.Green)

	colon := ":"
	for {
		pixelutils.Empty(bigPixel)
		if colon == ":" {
			colon = " "
		} else {
			colon = ":"
		}

		textPixel.Cls()
		fmt.Fprintf(textPixel, "%02d%s%02d", time.Now().Hour(), colon, time.Now().Minute())
		c <- pixel
		time.Sleep(500 * time.Millisecond)
	}
}
