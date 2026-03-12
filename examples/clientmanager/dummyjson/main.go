package main

import (
	"context"
	"fmt"

	"examples/clientmanager/dummyjson/product"
)

func main() {
	productRepository := product.NewHTTPRepository()
	ctx := context.Background()

	fmt.Println(productRepository.List(ctx))
	fmt.Println(productRepository.Create(ctx, product.Product{
		Title: "My First Product",
		Price: 5000,
		Stock: 9,
	}))
}
