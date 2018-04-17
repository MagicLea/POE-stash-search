package main

import (
	"flag"
	"fmt"

	"github.com/Magiclea/poe-stash-search/backend/services/crawler"
	"github.com/gin-gonic/gin"
	mgo "gopkg.in/mgo.v2"
)

var (
	mongoURL   = flag.String("mongo_url", "localhost", "mongo DB URL")
	poeBaseURL = flag.String("poe_base_url", "http://api.pathofexile.com", "poe base URL")

	mgoSession = &mgo.Session{}
)

func main() {
	flag.Parse()

	setupMongo(*mongoURL)
	startCrawler()

	r := gin.Default()
	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.Run()
}

func setupMongo(url string) {
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	mgoSession = session
	session.SetMode(mgo.Monotonic, true)
}

func startCrawler() {
	crawler := crawler.New(*poeBaseURL, mgoSession)
	go func() {
		// TODO: retrieve lastID from mysql db
		lastID := "2811-4457-4108-4795-1510"
		if err := crawler.FollowStashStream(lastID); err != nil {
			fmt.Println("crawler.FollowStashStream failed:", err)
		}
	}()
}
