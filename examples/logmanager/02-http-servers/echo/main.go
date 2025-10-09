package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"examples/logmanager/shared/models"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmecho"
	"github.com/labstack/echo/v4"
)

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("http-echo"),
	)

	userRepo := NewUserRepository()
	userService := NewUserService(userRepo)

	e := echo.New()
	e.Use(lmecho.Middleware(app))

	e.GET("/", healthCheck)
	e.GET("/users", getUsersHandler(userService))

	fmt.Println("Echo server running at http://localhost:8002")
	e.Logger.Fatal(e.Start(":8002"))
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func getUsersHandler(service *UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		users, err := service.GetAll(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": users,
		})
	}
}

type UserRepository struct {
	users  map[int]*models.User
	nextID int
	mu     sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: map[int]*models.User{
			1: {ID: 1, Name: "John Doe", Email: "john@example.com"},
			2: {ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
		},
		nextID: 3,
	}
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	txn := logmanager.StartDatabaseSegment(
		logmanager.FromContext(ctx),
		logmanager.DatabaseSegment{
			Table: "users",
			Query: "SELECT * FROM users",
		},
	)
	defer txn.End()

	time.Sleep(50 * time.Millisecond)

	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAll(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAll(ctx)
}