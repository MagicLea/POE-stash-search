package main

import (
	"fmt"

	"github.com/Magiclea/poe-stash-search/backend/services/crawler"
	"github.com/gin-gonic/gin"
)

func init() {
	crawler := crawler.New("http://api.pathofexile.com")
	go func() {
		// TODO: retrieve lastID from mysql db
		lastID := "2811-4457-4108-4795-1510"
		if err := crawler.FollowStashStream(lastID); err != nil {
			fmt.Println("crawler.FollowStashStream failed:", err)
		}
	}()
}

func main() {
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.Run()
}
