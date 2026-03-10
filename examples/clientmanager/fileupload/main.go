package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
)

type Request struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Response struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
}

func main() {
	req := &Request{
		Title:       "My Cool Photo",
		Description: "A cloudy lake",
	}
	res, err := clientmanager.Call[Response](
		context.Background(),
		"/post",
		clientmanager.WithRequestBody(req),
		clientmanager.WithMethod(http.MethodPost),
		clientmanager.WithHost("https://httpbin.org"),
		clientmanager.WithFiles(map[string]string{
			"image": "example.jpg",
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status:", res.StatusCode)
	fmt.Println("Uploaded:", res.Body.URL)
	fmt.Println("Success:", res.IsSuccess())
}
