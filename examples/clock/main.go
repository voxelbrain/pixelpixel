package main

import (
	"fmt"
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"time"
)

func colonGenerator() <-chan string {
	c := make(chan string)
	go func() {
		for {
			c <- ":"
			c <- " "
		}
	}()
	return c
}

func main() {
	wall, _ := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()
	bigPixel := pixelutils.DimensionChanger(pixel, 5*4, 18)
	textPixel := pixelutils.NewImageWriter(bigPixel, pixelutils.Green)

	colons := colonGenerator()
	for {
		pixelutils.Empty(bigPixel)
		textPixel.Cls()
		fmt.Fprintf(textPixel, "%02d%s%02d", time.Now().Hour(), <-colons, time.Now().Minute())
		wall <- pixel
		time.Sleep(500 * time.Millisecond)
	}
}
