package main

import (
	"context"
	"log"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
)

func main() {
	res, err := clientmanager.Call[any](
		context.Background(),
		"https://httpbin.org/basic-auth/user123/pass123",
		clientmanager.WithAuth(clientmanager.AuthBasic("user123", "pass123")),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res.IsSuccess())
}
