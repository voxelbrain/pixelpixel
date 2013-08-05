package twitter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/araddon/goauth"
	"github.com/araddon/httpstream"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type TwitterUser struct {
	Name           string
	ScreenName     string
	ProfilePicture image.Image
}

type Tweet struct {
	Text   string
	Date   time.Time
	Author TwitterUser
}

type Credentials struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	AccessToken    string `json:"access_token"`
	AccessSecret   string `json:"access_secret"`
}

func Hashtags(cred *Credentials, Hashtags ...string) (<-chan *Tweet, error) {
	c := make(chan *Tweet)
	restclient := twittergo.NewClient(&oauth1a.ClientConfig{
		ConsumerKey:    cred.ConsumerKey,
		ConsumerSecret: cred.ConsumerSecret,
	}, &oauth1a.UserConfig{
		AccessTokenKey:    cred.AccessToken,
		AccessTokenSecret: cred.AccessSecret,
	})

	httpstream.OauthCon = &oauth.OAuthConsumer{
		ConsumerKey:    cred.ConsumerKey,
		ConsumerSecret: cred.ConsumerSecret,
	}
	streamclient := httpstream.NewOAuthClient(&oauth.AccessToken{
		Token:  cred.AccessToken,
		Secret: cred.AccessSecret,
	}, newChannelConverter(c, restclient))

	err := streamclient.Filter(nil, Hashtags, nil, nil, false, nil)
	if err != nil {
		close(c)
	}
	return c, err
}

func newChannelConverter(c chan *Tweet, client *twittergo.Client) func([]byte) {
	rateLimitActive := false
	timer := time.AfterFunc(10*time.Second, func() {
		rateLimitActive = false
	})
	timer.Stop()

	return func(data []byte) {
		if rateLimitActive {
			return
		}

		rateLimitActive = true
		timer.Reset(10 * time.Second)

		var tweet twittergo.Tweet
		err := json.Unmarshal(data, &tweet)
		if err != nil {
			log.Printf("Received invalid tweet: %s", err)
			return
		}

		profileImageUrl, err := getUserProfileImageURL(client, tweet.User().Id())
		if err != nil {
			log.Printf("Could not get users profile image url: %s", err)
			return
		}
		profileImage, err := getImage(profileImageUrl)
		if err != nil {
			log.Printf("Could not get users profile image: %s", err)
			return
		}
		c <- &Tweet{
			Text: tweet.Text(),
			Date: tweet.CreatedAt(),
			Author: TwitterUser{
				ScreenName:     tweet.User().ScreenName(),
				Name:           tweet.User().Name(),
				ProfilePicture: profileImage,
			},
		}
	}
}

type user struct {
	ProfileImageUrl string `json:"profile_image_url"`
}

func getUserProfileImageURL(client *twittergo.Client, userId uint64) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/users/show.json?user_id=%d", userId), nil)
	if err != nil {
		return "", err
	}

	resp, err := client.SendRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var u user
	err = json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		return "", err
	}
	return u.ProfileImageUrl, nil
}

func getImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	return img, err
}
