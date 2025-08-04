package main

import (
	"context"
	"fmt"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager/examples/dummyjson/product"
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
