package product

import (
	"context"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
)

type httpRepository struct {
	clientManager clientmanager.ClientManager[Response]
}

func (r httpRepository) List(ctx context.Context) ([]Product, error) {
	res, err := r.clientManager.Call(ctx, "/products")
	if err != nil {
		return nil, err
	}
	return res.Body.Products, nil
}

func (r httpRepository) Create(ctx context.Context, p Product) error {
	req := &Request{
		Title: p.Title,
		Price: p.Price,
		Stock: p.Stock,
	}
	if _, err := r.clientManager.Call(
		ctx,
		"/products/add",
		clientmanager.WithRequestBody(req),
		clientmanager.WithMethod(http.MethodPost),
	); err != nil {
		return err
	}
	return nil
}

func NewHTTPRepository() Repository {
	return &httpRepository{clientmanager.New[Response](
		clientmanager.WithHost("https://dummyjson.com"),
	)}
}
