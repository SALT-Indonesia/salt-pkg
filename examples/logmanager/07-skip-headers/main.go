package main

import (
	"fmt"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const traceIDKey = "xid"

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	app := logmanager.NewApplication(
		logmanager.WithAppName("skip-headers-example"),
		logmanager.WithDebug(),
		logmanager.WithSkipHeaders(),
	)

	r.Use(traceIDMiddleware(), lmgin.Middleware(app))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	fmt.Println("Server running at http://localhost:8007")
	fmt.Println("Try: curl http://localhost:8007/ping")
	fmt.Println("Notice: request headers will NOT appear in the logs (WithSkipHeaders enabled)")

	if err := r.Run(":8007"); err != nil {
		panic(err)
	}
}

func traceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(traceIDKey, uuid.NewString())
		c.Next()
	}
}
