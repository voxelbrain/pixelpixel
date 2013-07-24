package main

import (
	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
	"image"
	_ "image/jpeg"
	"log"
	"net/http"
)

const (
	URL = `http://i.imgur.com/0VEDr0t.jpg`
)

func main() {
	c := protocol.PixelPusher()
	img := protocol.NewPixel()

	bigImg := downloadImage()
	imageutils.Copy(img, bigImg, bigImg.Bounds(), img.Bounds())
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
