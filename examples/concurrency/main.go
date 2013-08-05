package main

import (
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"image"
	"image/color"
	"image/draw"
	"time"
)

func main() {
	pusher := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	sub1 := pixelutils.SubImage(pixel, image.Rect(0, 0, 128, 256))
	sub2 := pixelutils.SubImage(pixel, image.Rect(128, 0, 256, 256))

	c := make(chan draw.Image)
	go Blinker(c, sub1, 901*time.Millisecond, pixelutils.Red, pixelutils.Green)
	go Blinker(c, sub2, 307*time.Millisecond, pixelutils.Yellow, pixelutils.Cyan, pixelutils.Magenta)
	go func() {
		for _ = range c {
			pusher <- pixel
		}
	}()

	// Block indefinitely
	select {}
}

func Blinker(c chan<- draw.Image, pixel draw.Image, sleep time.Duration, colors ...color.Color) {
	for {
		for _, color := range colors {
			pixelutils.Fill(pixel, color)
			c <- pixel
			time.Sleep(sleep)
		}
	}
}
