package main

import (
	"context"
	"fmt"
	"log"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
)

type RSS struct {
	Channel struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
	} `xml:"channel"`
}

func main() {
	res, err := clientmanager.Call[RSS](
		context.Background(),
		"/news/rss.xml",
		clientmanager.WithHost("https://feeds.bbci.co.uk"),
	) // get XML as string
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Feed Title:", res.Body.Channel.Title)
	fmt.Println("Feed Description:", res.Body.Channel.Description)
}
