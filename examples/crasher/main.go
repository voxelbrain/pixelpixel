package main

import (
	"image/color"
	"time"

	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
)

func main() {
	c := protocol.PixelPusher()
	img := protocol.NewPixel()

	dImg := imageutils.DimensionChanger(img, 4, 1)
	for i := 0; i < 5; i++ {
		if i < 3 {
			dImg.Set(i, 0, color.RGBA{uint8(100 + i*70), 0, 0, 255})
		} else if i == 3 {
			dImg.Set(i, 0, color.RGBA{0, 255, 0, 255})
		} else {
			panic("CRASH")
		}
		c <- img
		time.Sleep(1 * time.Second)
	}
}
