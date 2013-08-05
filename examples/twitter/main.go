package main

import (
	"encoding/json"
	"fmt"
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"github.com/voxelbrain/pixelpixel/pixelutils/twitter"
	"log"
	"os"
)

func main() {
	cred, err := loadCredentials()
	if err != nil {
		log.Fatalf("Could not load credentials: %s", err)
	}
	log.Printf("%#v", cred)

	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()
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
