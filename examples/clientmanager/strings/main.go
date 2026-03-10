package main

import (
	"context"
	"fmt"
	"log"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
)

func main() {
	res, err := clientmanager.Call[string](
		context.Background(),
		"/robots.txt",
		clientmanager.WithHost("https://httpbin.org"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Raw text:", res.Body)
	fmt.Println("Success:", res.IsSuccess())
}
