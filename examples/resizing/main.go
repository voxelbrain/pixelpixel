package main

import (
	"fmt"
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"image"
	_ "image/jpeg"
	"log"
	"net/http"
	"time"
)

const (
	URL = `http://i.imgur.com/0VEDr0t.jpg`
)

var (
	translucentBlack = pixelutils.Black
)

func main() {
	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	bigImage := downloadImage()

	start := time.Now()
	pixelutils.Resize(pixel, bigImage)
	convDuration := time.Now().Sub(start)

	textArea := pixelutils.SubPixel(pixel, pixel.Bounds().Intersect(image.Rect(0, 220, 999, 999)).Inset(8))
	pixelutils.Fill(textArea, translucentBlack)
	textArea = pixelutils.DimensionChanger(textArea, 60, 6).(pixelutils.Pixel)
	text := fmt.Sprintf("Conv: %s", convDuration)
	pixelutils.DrawText(textArea, pixelutils.Red, text)

	c <- pixel
	select {}
}

func downloadImage() image.Image {
	resp, err := http.Get(URL)
	if err != nil {
		log.Fatalf("Could not obtain image: %s", err)
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Fatalf("Could not decoe image: %s", err)
	}
	return img
}

func init() {
	translucentBlack.A = 150
}
