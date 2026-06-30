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

	// Production mode: WithEnvironment("production") forces debug=false (no
	// WithDebug()). Even so, the headers matched by the WithExposeHeaders
	// wildcards below are still logged - infrastructure headers from Cloudflare
	// (CF-*) and CloudFront (X-Amz-Cf-*, X-Amzn-*) survive, everything else is
	// dropped. Wildcard matching is case-insensitive, so "CF-*" matches the
	// canonicalized "Cf-Ray" Go produces.
	app := logmanager.NewApplication(
		logmanager.WithAppName("expose-header-prefixes-example"),
		logmanager.WithEnvironment("production"),
		logmanager.WithExposeHeaders("CF-*", "X-Amz-Cf-*", "X-Amzn-*", "X-Request-Id"),
	)

	r.Use(traceIDMiddleware(), lmgin.Middleware(app))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	fmt.Println("Server running at http://localhost:8008 (production mode, debug=false)")
	fmt.Println("Try:")
	fmt.Println(`  curl http://localhost:8008/ping \`)
	fmt.Println(`    -H "CF-Ray: 8abc" \`)
	fmt.Println(`    -H "CF-Connecting-IP: 1.2.3.4" \`)
	fmt.Println(`    -H "X-Amzn-Trace-Id: Root=1-xyz" \`)
	fmt.Println(`    -H "X-Amz-Cf-Id: zzz" \`)
	fmt.Println(`    -H "User-Agent: noise"`)
	fmt.Println("Notice: only the CF-*, X-Amz-Cf-*, X-Amzn-* (and X-Request-Id) headers")
	fmt.Println("appear in the logged \"headers\" object - User-Agent is dropped - even though debug is off.")

	if err := r.Run(":8008"); err != nil {
		panic(err)
	}
}

func traceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(traceIDKey, uuid.NewString())
		c.Next()
	}
}
