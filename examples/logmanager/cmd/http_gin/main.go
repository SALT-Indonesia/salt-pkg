package main

import (
	"context"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"time"
)

const contextKey = "xid"

func middlewareTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(contextKey, uuid.NewString())
	}
}

func main() {
	r := gin.Default()

	app := logmanager.NewApplication(
		logmanager.WithAppName("http-gin"),
		logmanager.WithTraceIDContextKey(contextKey),
		logmanager.WithExposeHeaders("X-Forwarded-For", "X-Url-Payload"), // this is for print request headers
		//logmanager.WithTraceIDHeaderKey("X-Trace-Id"),
		//logmanager.WithLogDir("~/GolandProjects/go-standard-log/_local"),
		//logmanager.WithTraceIDKey("xid"), // (optional) add xid if you want to change key of trace id
	)

	r.Use(middlewareTraceID(), lmgin.Middleware(app))

	r.POST("/register", func(c *gin.Context) {
		// tx := logmanager.FromContext(c)
		// tx.SkipResponse()
		// tx.SkipResponse()

		_ = registerHandler(c)

		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	r.POST("/event", func(c *gin.Context) {
		title := c.PostForm("title")
		description := c.PostForm("description")
		location := c.PostForm("location")

		file, err := c.FormFile("poster")
		var fileInfo map[string]interface{}
		if err == nil && file != nil {
			fileInfo = map[string]interface{}{
				"filename": file.Filename,
				"size":     file.Size,
			}
		}

		c.JSON(200, gin.H{
			"status":  201,
			"message": "event created successfully",
			"event": gin.H{
				"title":       title,
				"description": description,
				"location":    location,
				"file":        fileInfo,
			},
		})
	})

	fmt.Println("Server is running at :8000")
	panic(r.Run(":8000"))
}

func registerHandler(ctx context.Context) error {
	txn := logmanager.StartOtherSegment(
		logmanager.FromContext(ctx),
		logmanager.OtherSegment{
			Name: "segment",
		},
	)
	defer txn.End()

	time.Sleep(230 * time.Millisecond)
	return nil
}
