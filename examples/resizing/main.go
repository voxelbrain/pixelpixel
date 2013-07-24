package main

import (
	"fmt"
	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
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
	c := protocol.PixelPusher()
	img := protocol.NewPixel()

	start := time.Now()
	bigImg := downloadImage()
	dlDuration := time.Now().Sub(start)

	start = time.Now()
	imageutils.Copy(img, bigImg, bigImg.Bounds(), img.Bounds())
	convDuration := time.Now().Sub(start)

	textImg := imageutils.DimensionChanger(img, 128, 128)
	r := textImg.Bounds()
	text := fmt.Sprintf("DL:   %s\nConv: %s", dlDuration, convDuration)
	imageutils.DrawText(textImg, image.Rect(r.Min.X, r.Max.Y-12, r.Max.X, r.Max.Y), imageutils.Red, text)
	c <- img
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
