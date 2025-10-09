package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const traceIDKey = "xid"

func main() {
	r := gin.Default()

	app := logmanager.NewApplication(
		logmanager.WithAppName("http-gin"),
		logmanager.WithTraceIDContextKey(traceIDKey),
		logmanager.WithExposeHeaders("X-Forwarded-For", "X-Url-Payload"),
	)

	r.Use(traceIDMiddleware(), lmgin.Middleware(app))

	r.POST("/register", registerHandler)

	fmt.Println("Gin server running at http://localhost:8001")
	if err := r.Run(":8001"); err != nil {
		panic(err)
	}
}

func traceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(traceIDKey, uuid.NewString())
		c.Next()
	}
}

func registerHandler(c *gin.Context) {
	processRegistration(c)

	c.JSON(200, gin.H{
		"message": "registration successful",
	})
}

func processRegistration(ctx context.Context) {
	txn := logmanager.StartOtherSegment(
		logmanager.FromContext(ctx),
		logmanager.OtherSegment{
			Name: "process-registration",
		},
	)
	defer txn.End()

	time.Sleep(230 * time.Millisecond)
}