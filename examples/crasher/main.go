package main

import (
	"image/color"
	"time"

	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
)

func main() {
	protocol.ServePixel(func(p *protocol.Pixel) {
		img := imageutils.DimensionChanger(p, 4, 1)
		for i := 0; i < 5; i++ {
			if i < 3 {
				img.Set(i, 0, color.RGBA{uint8(100 + i*70), 0, 0, 255})
			} else if i == 3 {
				img.Set(i, 0, color.RGBA{0, 255, 0, 255})
			} else {
				panic("CRASH")
			}
			p.Commit()
			time.Sleep(1 * time.Second)
		}
	})
}
