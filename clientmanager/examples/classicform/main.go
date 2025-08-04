package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
)

type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Message string `json:"message"`
}

func main() {
	req := &Request{
		Username: "alice",
		Password: "s3cr3t",
	}
	res, err := clientmanager.Call[Response](
		context.Background(),
		"/post",
		clientmanager.WithRequestBody(req),
		clientmanager.WithMethod(http.MethodPost),
		clientmanager.WithHost("https://httpbin.org"),
		clientmanager.WithFormURLEncoded(),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status Code:", res.StatusCode)
	fmt.Println("Raw:", string(res.Raw))
	fmt.Println("Success:", res.IsSuccess())
}
