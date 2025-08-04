package main

import (
	"context"
	"log"

	"github.com/SALT-Indonesia/salt-pkg/eventmanager"
	"github.com/joho/godotenv"
)

type message struct {
	Number      int    `json:"number"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	// for i := range 10 {
	myMessage := &message{
		// Number:      i,
		Title:       "This is my message",
		Description: "This is the description of my message.",
	}
	if err := eventmanager.Publish(context.Background(), "mytopic", myMessage); err != nil {
		log.Fatal(err)
	}
	// }
}
