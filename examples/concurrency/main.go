package main

import (
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"image"
	"image/color"
	"time"
)

var (
	c     = pixelutils.PixelPusher()
	pixel = pixelutils.NewPixel()
)

func main() {

	sub1 := pixelutils.SubPixel(pixel, image.Rect(0, 0, 128, 256))
	sub2 := pixelutils.SubPixel(pixel, image.Rect(128, 0, 256, 256))

	go Blinker(sub1, 901*time.Millisecond, pixelutils.Red, pixelutils.Green)
	go Blinker(sub2, 307*time.Millisecond, pixelutils.Yellow, pixelutils.Cyan, pixelutils.Magenta)

	// Block indefinitely
	select {}
}

func Blinker(spixel pixelutils.Pixel, sleep time.Duration, colors ...color.Color) {
	for {
		for _, color := range colors {
			pixelutils.Fill(spixel, color)
			c <- pixel
			time.Sleep(sleep)
		}
	}
}
