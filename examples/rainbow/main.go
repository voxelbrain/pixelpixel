package main

import (
	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
)

func main() {
	protocol.ServePixel(func(p *protocol.Pixel) {
		for y := 0; y < 256; y++ {
			for x := 0; x < 256; x++ {
				c := imageutils.HSLA{
					H: float64(x) / 256,
					S: 1.0,
					L: 0.5 * (1 - float64(y)/256),
					A: 1.0,
				}
				p.Set(x, y, c)
			}
		}
		p.Commit()
		// Block indefinitely
		select {}
	})
}
