package main

import (
	"context"
)

type UserUsecase interface {
	GetAllUsers(ctx context.Context) ([]*User, error)
}

type UserUsecaseImpl struct {
	userRepo UserRepository
}

func NewUserUsecase(userRepo UserRepository) UserUsecase {
	return &UserUsecaseImpl{
		userRepo: userRepo,
	}
}

func (u *UserUsecaseImpl) GetAllUsers(ctx context.Context) ([]*User, error) {
	return u.userRepo.GetAll(ctx)
}
