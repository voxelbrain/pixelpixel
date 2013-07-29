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

func main() {
	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	bigImage := downloadImage()

	start := time.Now()
	pixelutils.Copy(pixel, bigImage, bigImage.Bounds(), pixel.Bounds())
	convDuration := time.Now().Sub(start)

	textArea := pixelutils.DimensionChanger(pixel, 64, 64)
	r := textArea.Bounds()
	text := fmt.Sprintf("Conv: %s", convDuration)
	pixelutils.DrawText(textArea, image.Rect(r.Min.X, r.Max.Y-6, r.Max.X, r.Max.Y), pixelutils.Red, text)
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
