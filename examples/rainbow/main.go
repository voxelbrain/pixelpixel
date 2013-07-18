package main

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/voxelbrain/pixelpixel/protocol"
)

var (
	red    = color.NRGBA{255, 0, 0, 255}
	green  = color.NRGBA{0, 255, 0, 255}
	blue   = color.NRGBA{0, 0, 255, 255}
	colors = []color.Color{red, green, blue}
)

func main() {
	time.AfterFunc(4*time.Second, func() {
		panic("ah")
	})
	protocol.ServePixel(func(p *protocol.Pixel) {
		for {
			for _, fillColor := range colors {
				draw.Draw(p, image.Rect(0, 0, 256, 256), &image.Uniform{fillColor}, image.Point{0, 0}, draw.Over)
				p.Commit()
				time.Sleep(1000 * time.Millisecond)
			}
		}
	})
}
