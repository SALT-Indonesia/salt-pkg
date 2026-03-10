package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
)

type Request struct {
	Title string  `json:"title" validate:"required"`
	Price float64 `json:"price" validate:"required"`
}

type Response struct {
	ID uint64 `json:"id"`
}

func main() {
	req := &Request{
		Title: "My Product",
		Price: 123.45,
	}
	res, err := clientmanager.Call[Response](
		context.Background(),
		"/post",
		clientmanager.WithRequestBody(req),
		clientmanager.WithMethod(http.MethodPost),
		clientmanager.WithHost("https://httpbin.org"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("HTTP Status:", res.StatusCode)
	fmt.Println("Response Struct:", res.Body)
	fmt.Println("Raw Body:", string(res.Raw))
	fmt.Println("Success:", res.IsSuccess())
}
