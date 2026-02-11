package main

import (
	"fmt"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmecho"
	"github.com/labstack/echo/v4"
)

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("echo-multipart-example"),
		logmanager.WithDebug(),
		logmanager.WithSplitLevelOutput(),
	)

	e := echo.New()
	e.Use(lmecho.Middleware(app))

	e.GET("/", healthCheck)
	e.POST("/upload", uploadHandler)
	e.POST("/contact", contactFormHandler)

	fmt.Println("Echo multipart form-data example server")
	fmt.Println("=========================================")
	fmt.Println("")
	fmt.Println("Endpoints:")
	fmt.Println("  GET  http://localhost:8003/")
	fmt.Println("  POST http://localhost:8003/upload")
	fmt.Println("  POST http://localhost:8003/contact")
	fmt.Println("")
	fmt.Println("Try these commands:")
	fmt.Println("  curl http://localhost:8003/")
	fmt.Println("")
	fmt.Println("  # Form data only:")
	fmt.Println("  curl -X POST http://localhost:8003/contact -F \"name=John Doe\" -F \"email=john@example.com\" -F \"message=Hello!\"")
	fmt.Println("")
	fmt.Println("  # File upload with form fields:")
	fmt.Println("  curl -X POST http://localhost:8003/upload -F \"title=My Document\" -F \"file=@README.md\"")
	fmt.Println("")

	e.Logger.Fatal(e.Start(":8003"))
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func contactFormHandler(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	message := c.FormValue("message")

	if name == "" || email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name and email are required",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "received",
		"name":    name,
		"email":   email,
		"message": message,
	})
}

func uploadHandler(c echo.Context) error {
	title := c.FormValue("title")

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "file is required",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "uploaded",
		"title":    title,
		"filename": file.Filename,
		"size":     file.Size,
	})
}
