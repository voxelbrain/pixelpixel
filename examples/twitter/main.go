package main

import (
	"encoding/json"
	"fmt"
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"github.com/voxelbrain/pixelpixel/pixelutils/twitter"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"log"
	"os"
)

type TweetSection struct {
	BackgroundImage string
	Hashtag         string
}

var (
	TweetSections = []TweetSection{
		{
			BackgroundImage: `imgs/windows.png`,
			Hashtag:         "#Windows",
		},
		// {
		// 	BackgroundImage: `imgs/osx.png`,
		// 	Hashtag:         "#OSX",
		// },
		// {
		// 	BackgroundImage: `imgs/ubuntu.png`,
		// 	Hashtag:         "#Ubuntu",
		// },
	}
	translucentBlack = color.RGBA{0, 0, 0, 230}
)

func main() {
	cred, err := loadCredentials()
	if err != nil {
		log.Fatalf("Could not load credentials: %s", err)
	}

	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	fakeC := make(chan draw.Image)
	for i, section := range TweetSections {
		subPixel := pixelutils.SubImage(pixel, image.Rect(0, i*85, 256, (i+1)*85))

		pixelutils.Resize(subPixel, loadImage(section.BackgroundImage))
		pixelutils.Fill(subPixel, translucentBlack)
		tweets, err := twitter.Hashtags(cred, section.Hashtag)
		if err != nil {
			log.Printf("Could not start tweet streamer for hashtag \"%s\": %s", section.Hashtag, err)
			continue
		}
		go TweetDrawer(fakeC, subPixel, tweets)

	}
	for _ = range fakeC {
		c <- pixel
	}
}

func TweetDrawer(c chan<- draw.Image, pixel draw.Image, tweets <-chan *twitter.Tweet) {
	bg := image.NewNRGBA(pixel.Bounds())
	pixelutils.Resize(bg, pixel)
	c <- pixel
	avatarArea := pixelutils.SubImage(pixel, image.Rectangle{
		Min: pixel.Bounds().Min,
		Max: pixel.Bounds().Min.Add(image.Point{85, 85}),
	})
	textArea := pixelutils.PixelSizeChanger(pixelutils.SubImage(pixel, image.Rectangle{
		Min: pixel.Bounds().Min.Add(image.Point{85, 0}),
		Max: pixel.Bounds().Max,
	}), 2, 2)
	for tweet := range tweets {
		pixelutils.Resize(pixel, bg)
		pixelutils.Resize(avatarArea, tweet.Author.ProfilePicture)
		pixelutils.DrawText(textArea, pixelutils.White, tweet.Text)
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
