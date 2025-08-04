package application

import "context"

type RegisterInput struct {
	Name string
}

type RegisterOutput struct {
	Message string
}

type RegisterUseCase interface {
	Execute(ctx context.Context, input RegisterInput) (RegisterOutput, error)
}
