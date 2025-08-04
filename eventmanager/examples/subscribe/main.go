package main

import (
	"context"
	"log"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/eventmanager"
	"github.com/joho/godotenv"
)

type message struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func handle(m message) (domainErr, infrastructureErr error) {
	time.Sleep(1 * time.Second)
	// if m.Number == 5 {
	// 	return nil, errors.New("failed infrastructure")
	// }
	log.Println(m)
	return nil, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	eventmanager.Subscribe(context.Background(), "myservice", "mytopic", []eventmanager.Handler[message]{handle})
}
