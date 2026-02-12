package main

import (
	"fmt"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logmanager application
	app := logmanager.NewApplication(
		logmanager.WithAppName("gin-async-test"),
		logmanager.WithDebug(),
	)

	// Create Gin router
	r := gin.Default()

	// Add logmanager middleware
	r.Use(lmgin.Middleware(app))

	// Endpoint that simulates async processing
	r.GET("/async", func(c *gin.Context) {
		// Get transaction from context
		tx := logmanager.FromContext(c.Request.Context())

		// Simulate database operation (synchronous)
		time.Sleep(10 * time.Millisecond)

		// Spawn goroutine for async processing (this is where the bug occurs)
		go func() {
			// Clone context for goroutine (this is the pattern users typically follow)
			ctx := c.Copy()

			// Simulate async work (e.g., calling external API, background processing)
			time.Sleep(100 * time.Millisecond)

			// Try to use transaction in async context
			// Before the fix, this would panic because tx.attrs has been set to nil
			// after the main handler returned and the middleware called tx.End()
			if tx != nil {
				// This would cause panic: invalid memory address or nil pointer dereference
				tx.SetWebResponse(logmanager.WebResponse{
					StatusCode: 200,
					Body:       []byte(`{"async": "completed"}`),
				})
			}

			fmt.Printf("Async work completed in goroutine with context: %v\n", ctx.Request.URL.Path)
		}()

		// Handler returns immediately (before goroutine completes)
		c.JSON(200, gin.H{
			"message": "Request received, processing asynchronously",
		})
	})

	// Endpoint to test high concurrency
	r.GET("/concurrent", func(c *gin.Context) {
		tx := logmanager.FromContext(c.Request.Context())

		// Spawn multiple goroutines to simulate high traffic scenario
		for i := 0; i < 10; i++ {
			go func(id int) {
				time.Sleep(time.Duration(id*10) * time.Millisecond)

				// All goroutines trying to access transaction concurrently
				if tx != nil {
					tx.SetWebResponse(logmanager.WebResponse{
						StatusCode: 200,
						Body:       []byte(fmt.Sprintf(`{"goroutine": %d}`, id)),
					})
				}
			}(i)
		}

		c.JSON(200, gin.H{
			"message": "Spawned 10 concurrent goroutines",
		})
	})

	// Simple health check endpoint (no async, should never fail)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Start server
	fmt.Println("Starting server on :8080")
	fmt.Println("Test endpoints:")
	fmt.Println("  - GET /async       - Test single async goroutine")
	fmt.Println("  - GET /concurrent  - Test multiple concurrent goroutines")
	fmt.Println("  - GET /health      - Health check (no async)")

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
