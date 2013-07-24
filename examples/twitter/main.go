package main

import (
	"fmt"
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"github.com/voxelbrain/pixelpixel/pixelutils/twitter"
	"log"
)

var (
	cred = &twitter.Credentials{
		ConsumerKey:    "aFtD6fImgT75BuKl1Cq9bQ",
		ConsumerSecret: "MYEuIGgdfo8RJjFOgcyYRTlxKb4v6NMgvcCV1virXs",
		AccessToken:    "15180856-KwFLdIhcYrqH9BNfo0FlhlAvTURzkwpmuyPv4TvbF",
		AccessSecret:   "go5PAVEC3eYCai5uRm0QvpPFoE67zh8XR6ooPhg",
	}
)

func main() {
	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()
	tweets := twitter.Hashtags(cred, "#chromecast")
	for tweet := range tweets {
		log.Printf("%#v", tweet)
		pixelutils.FillRectangle(pixel, pixel.Bounds(), pixelutils.Black)
		pic := tweet.Author.ProfilePicture
		pixelutils.Copy(pixel, pic, pic.Bounds(), pixel.Bounds())
		text := fmt.Sprintf("%s (@%s):\n%s", tweet.Author.Name, tweet.Author.ScreenName, tweet.Text)
		pixelutils.DrawText(pixel, pixel.Bounds(), pixelutils.Red, text)
		c <- pixel
	}

}
