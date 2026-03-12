package product

import "context"

type Repository interface {
	List(context.Context) ([]Product, error)
	Create(context.Context, Product) error
}
