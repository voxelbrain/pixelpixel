package main

import (
	"image"
	"image/draw"
	"time"

	"github.com/voxelbrain/pixelpixel/protocol"
)

func main() {
	time.AfterFunc(4*time.Second, func() {
		panic("CRASH")
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
