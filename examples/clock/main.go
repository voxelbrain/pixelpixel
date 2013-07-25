package main

import (
	"fmt"
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"image"
	"log"
	"time"
)

func main() {
	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()
	img := pixelutils.DimensionChanger(pixel, 5*4, 18)

	colon := ":"
	for {
		pixelutils.Empty(img)
		if colon == ":" {
			colon = " "
		} else {
			colon = ":"
		}
		log.Printf("time=\"%s\", colon=\"%s\"", time.Now(), colon)
		timeStr := fmt.Sprintf("%02d%s%02d", time.Now().Hour(), colon, time.Now().Minute())
		pixelutils.DrawText(img, image.Rect(0, 6, 5*4, 12), pixelutils.Green, timeStr)
		c <- pixel
		time.Sleep(500 * time.Millisecond)
	}
}
