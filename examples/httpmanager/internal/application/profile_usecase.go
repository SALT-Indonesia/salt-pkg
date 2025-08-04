package application

import "context"

type ProfileOutput struct {
	Name string
}

type ProfileUseCase interface {
	Execute(ctx context.Context) (ProfileOutput, error)
}
