package main

import (
	"encoding/json"
	"fmt"
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"github.com/voxelbrain/pixelpixel/pixelutils/twitter"
	"image"
	_ "image/png"
	"log"
	"os"
)

const (
	windowsImage = `imgs/windows.png`
	osxImage     = `imgs/osx.png`
	ubuntuImage  = `imgs/ubuntu.png`
)

func main() {
	cred, err := loadCredentials()
	if err != nil {
		log.Fatalf("Could not load credentials: %s", err)
	}

	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()
	windowsPixel := pixelutils.SubImage(pixel, image.Rect(0, 0*85, 256, 1*85))
	osxPixel := pixelutils.SubImage(pixel, image.Rect(0, 1*85, 256, 2*85))
	ubuntuPixel := pixelutils.SubImage(pixel, image.Rect(0, 2*85, 256, 3*85))

	pixelutils.Resize(windowsPixel, loadImage(windowsImage))
	pixelutils.Resize(osxPixel, loadImage(osxImage))
	pixelutils.Resize(ubuntuPixel, loadImage(ubuntuImage))

	c <- pixel
	select {}

	tweets, err := twitter.Hashtags(cred, "#OSX")
	if err != nil {
		log.Fatalf("Could not open Twitter stream: %s", err)
	}
	for tweet := range tweets {
		pixelutils.Fill(pixel, pixelutils.Black)
		pic := tweet.Author.ProfilePicture
		pixelutils.Resize(pixel, pic)
		text := fmt.Sprintf("%s (@%s):\n%s", tweet.Author.Name, tweet.Author.ScreenName, tweet.Text)
		pixelutils.DrawText(pixel, pixelutils.Red, text)
		c <- pixel
	}

}

func loadImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Could not open hard-coded source image: %s", err))
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(fmt.Sprintf("Could not decode hard-coded source image: %s", err))
	}
	return img
}

func loadCredentials() (*twitter.Credentials, error) {
	f, err := os.Open("credentials.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cred *twitter.Credentials
	err = json.NewDecoder(f).Decode(&cred)
	return cred, err
}
