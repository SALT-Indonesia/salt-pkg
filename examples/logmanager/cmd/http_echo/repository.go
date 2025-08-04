package main

import (
	"context"
	"sync"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserRepository interface {
	GetAll(ctx context.Context) ([]*User, error)
}

type InMemoryUserRepository struct {
	users  map[int]*User
	nextID int
	mu     sync.RWMutex
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	repo := &InMemoryUserRepository{
		users:  make(map[int]*User),
		nextID: 1,
		mu:     sync.RWMutex{},
	}

	repo.users[1] = &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	repo.users[2] = &User{ID: 2, Name: "Jane Smith", Email: "jane@example.com"}
	repo.nextID = 3

	return repo
}

func (r *InMemoryUserRepository) GetAll(ctx context.Context) ([]*User, error) {
	txn := logmanager.StartDatabaseSegment(
		logmanager.FromContext(ctx),
		logmanager.DatabaseSegment{
			Table: "users",
			Query: "select * from users",
		},
	)
	defer txn.End()

	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}
