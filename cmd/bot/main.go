package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/Davincible/goinsta/v3"
	"github.com/nekoding/stikombali-instagram/internal/scrapper"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AccountInfoList struct {
	Username []string `yaml:"accounts"`
}

type Post struct {
	gorm.Model
	FeedID          string
	AccountName     string
	AccountUsername string
	Description     string `gorm:"type:longtext"`
	Media           string `gorm:"type:longtext;serializer:json"`
	Source          string
}

func saveToDatabase(feed scrapper.FeedInstagram, db *gorm.DB) {
	imgs, _ := json.Marshal(feed.Images)

	post := &Post{
		FeedID:          feed.FeedID,
		Description:     feed.Caption,
		Media:           string(imgs),
		Source:          "instagram",
		AccountName:     feed.Account.FullName,
		AccountUsername: feed.Account.Username,
	}

	db.FirstOrCreate(&post, Post{FeedID: post.FeedID})
}

func scrapeFeedInstagram(insta *goinsta.Instagram, account string, db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	feeds, err := scrapper.GetLatestFeed(insta, account)
	if err != nil {
		log.Panic(err)
	}

	for _, feed := range feeds {
		saveToDatabase(feed, db)
	}
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// migrate
	err = db.AutoMigrate(&Post{})

	if err != nil {
		panic(err)
	}

	insta, err := goinsta.Import(".authconfig")
	if err != nil {
		insta := goinsta.New(os.Getenv("IG_USERNAME"), os.Getenv("IG_PASSWORD"))
		if err := insta.Login(); err != nil {
			panic(err)
		}
	}

	defer insta.Export(".authconfig")

	configFile, err := os.ReadFile("config.yml")

	if err != nil {
		panic(err)
	}

	accounts := &AccountInfoList{}
	_ = yaml.Unmarshal(configFile, accounts)

	var wg sync.WaitGroup
	for _, account := range accounts.Username {
		wg.Add(1)
		go scrapeFeedInstagram(insta, account, db, &wg)
	}

	wg.Wait()
}
