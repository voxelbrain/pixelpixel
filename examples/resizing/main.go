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
	pixelutils.StretchCopy(pixel, bigImage)
	convDuration := time.Now().Sub(start)

	textArea := pixel.Bounds().Intersect(image.Rect(0, 220, 999, 999)).Inset(8)
	text := pixelutils.NewImageWriter(pixelutils.DimensionChanger(pixelutils.SubPixel(pixel, textArea), 60, 6), pixelutils.Red)

	pixelutils.Fill(text, translucentBlack)
	fmt.Fprintf(text, "Conv: %s", convDuration)

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
