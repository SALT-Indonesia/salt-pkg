package main

import (
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmecho"
	"github.com/labstack/echo/v4"
)

func main() {
	app := logmanager.NewApplication()

	userRepo := NewInMemoryUserRepository()
	userUsecase := NewUserUsecase(userRepo)

	e := echo.New()
	e.Use(lmecho.Middleware(app))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "ok",
		})
	})

	e.GET("/users", func(c echo.Context) error {
		users, err := userUsecase.GetAllUsers(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": users,
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
