package scrapper

import (
	"encoding/base64"

	"github.com/Davincible/goinsta/v3"
)

type AccountInfo struct {
	FullName string
	Username string
}

type FeedInstagram struct {
	Account AccountInfo
	FeedID  string
	Caption string
	Images  []string
}

func GetLatestFeed(insta *goinsta.Instagram, username string) ([]FeedInstagram, error) {
	var result []FeedInstagram

	p, err := insta.VisitProfile(username)
	if err != nil {
		return result, err
	}

	// follow account
	_ = p.User.Follow()

	feeds := p.Feed.Latest()
	for _, f := range feeds {
		var feed FeedInstagram
		var media []string

		images := f.CarouselMedia
		if len(images) > 0 {
			for _, image := range images {
				img, _ := image.Download()
				b64img := base64.StdEncoding.EncodeToString(img)

				media = append(media, b64img)
			}
		} else {
			img, _ := f.Download()
			b64img := base64.StdEncoding.EncodeToString(img)

			media = append(media, b64img)
		}

		feed.Account.FullName = p.User.FullName
		feed.Account.Username = p.User.Username
		feed.FeedID = f.GetID()
		feed.Caption = f.Caption.Text
		feed.Images = media

		result = append(result, feed)
	}

	return result, nil
}
