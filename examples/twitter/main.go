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
	"strings"
)

type TweetSection struct {
	BackgroundImage string
	Hashtag         string
	Tweets          chan *twitter.Tweet
}

var (
	TweetSections = []TweetSection{
		{
			BackgroundImage: `imgs/windows.png`,
			Hashtag:         "#Windows",
			Tweets:          make(chan *twitter.Tweet),
		},
		{
			BackgroundImage: `imgs/osx.png`,
			Hashtag:         "#OSX",
			Tweets:          make(chan *twitter.Tweet),
		},
		{
			BackgroundImage: `imgs/ubuntu.png`,
			Hashtag:         "#Ubuntu",
			Tweets:          make(chan *twitter.Tweet),
		},
	}
	translucentBlack = color.RGBA{0, 0, 0, 200}
)

func main() {
	cred := loadCredentials()
	fakeC := make(chan draw.Image)
	wall, _ := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	TweetDispatch(cred)

	for i, section := range TweetSections {
		subPixel := pixelutils.SubImage(pixel, image.Rect(0, i*85, 256, (i+1)*85))
		pixelutils.StretchCopy(subPixel, loadImage(section.BackgroundImage))
		pixelutils.Fill(subPixel, translucentBlack)
		go TweetDrawer(fakeC, subPixel, section.Tweets)
	}

	for _ = range fakeC {
		wall <- pixel
	}
}

func TweetDispatch(cred *twitter.Credentials) {
	allHashtags := make([]string, 0, len(TweetSections))
	for _, section := range TweetSections {
		allHashtags = append(allHashtags, section.Hashtag)
	}

	tweets, err := twitter.Hashtags(cred, allHashtags...)
	if err != nil {
		log.Fatalf("Could not start Twitter stream: %s", err)
	}
	go func() {
		for tweet := range tweets {
			for _, section := range TweetSections {
				if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(section.Hashtag)) {
					section.Tweets <- tweet
				}
			}
		}
	}()
}

func TweetDrawer(c chan<- draw.Image, pixel draw.Image, tweets <-chan *twitter.Tweet) {
	counter := 0
	bg := image.NewNRGBA(pixel.Bounds())
	pixelutils.StretchCopy(bg, pixel)
	c <- pixel
	avatarArea := pixelutils.SubImage(pixel, image.Rectangle{
		Min: pixel.Bounds().Min,
		Max: pixel.Bounds().Min.Add(image.Point{85, 85}),
	})
	textArea := pixelutils.PixelSizeChanger(pixelutils.SubImage(pixel, image.Rectangle{
		Min: pixel.Bounds().Min.Add(image.Point{90, 0}),
		Max: pixel.Bounds().Max,
	}), 2, 2)
	for tweet := range tweets {
		counter++
		pixelutils.StretchCopy(pixel, bg)
		pixelutils.StretchCopy(avatarArea, tweet.Author.ProfilePicture)
		pixelutils.DrawText(pixelutils.PixelSizeChanger(avatarArea, 3, 3), pixelutils.Red, fmt.Sprintf("%03d", counter))
		pixelutils.DrawText(textArea, pixelutils.White, tweet.Text)
		c <- pixel
	}
}

func loadImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Could not open hard-coded source image: %s", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("Could not decode hard-coded source image: %s", err)
	}
	return img
}

func loadCredentials() *twitter.Credentials {
	f, err := os.Open("credentials.json")
	if err != nil {
		log.Fatalf("Could not open credentials file: %s", err)
	}
	defer f.Close()

	var cred *twitter.Credentials
	err = json.NewDecoder(f).Decode(&cred)
	if err != nil {
		log.Fatalf("Could not decode credentials file: %s", err)
	}
	return cred
}
