package main

import (
	"image/color"
	"time"

	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
)

func main() {
	protocol.ServePixel(func(p *protocol.Pixel) {
		img := DonutImage{imageutils.DimensionChanger(p, 64, 64)}
		x, y := 5, 0
		for {
			img.Set(x, y, color.White)
			p.Commit()
			x, y = x+1, y+1
			time.Sleep(100 * time.Millisecond)
		}
	})
}
