package application

import (
	"context"
	"log"
	"time"
)

type registerUseCaseImpl struct {
}

func (u *registerUseCaseImpl) Execute(_ context.Context, input RegisterInput) (RegisterOutput, error) {
	log.Println("Executing usecase: ", input.Name)
	time.Sleep(200 * time.Millisecond)
	return RegisterOutput{Message: "Hello " + input.Name}, nil
}

func NewUseCaseImpl() RegisterUseCase {
	return &registerUseCaseImpl{}
}

type profileUseCaseImpl struct {
}

func (u *profileUseCaseImpl) Execute(_ context.Context) (ProfileOutput, error) {
	log.Println("Executing usecase: ")
	time.Sleep(200 * time.Millisecond)
	return ProfileOutput{Name: "John Doe"}, nil
}

func NewProfileUseCaseImpl() ProfileUseCase {
	return &profileUseCaseImpl{}
}
