package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create logmanager application with OpenTelemetry enabled
	app := logmanager.NewApplication(
		logmanager.WithService("otel-example-service"),
		logmanager.WithEnvironment("development"),
		logmanager.WithDebug(),
		// Enable OpenTelemetry trace export
		logmanager.WithOpenTelemetry(
			logmanager.WithOTelEndpoint("localhost:4317"),
			logmanager.WithOTelInsecure(),
		),
	)

	// Setup Gin router with logmanager middleware
	router := gin.New()
	router.Use(lmgin.Middleware(app))

	// Register handlers
	router.GET("/api/users", getUsers)
	router.GET("/api/users/:id", getUser)
	router.POST("/api/orders", createOrder)

	// Start server
	fmt.Println("Server starting on :8080")
	fmt.Println("OpenTelemetry traces will be exported to localhost:4317")
	fmt.Println("\nMake some requests to generate traces:")
	fmt.Println("  curl http://localhost:8080/api/users")
	fmt.Println("  curl http://localhost:8080/api/users/123")
	fmt.Println("  curl -X POST http://localhost:8080/api/orders -H 'Content-Type: application/json' -d '{\"user_id\": 123}'")
	fmt.Println("\nView traces in Jaeger: http://localhost:16686")
	fmt.Println("Make sure Jaeger is running: docker run -d -p 4317:4317 -p 16686:16686 jaegertracing/all-in-one:latest")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getUsers returns a list of users with database simulation
func getUsers(c *gin.Context) {
	// Get transaction from context
	tx := logmanager.FromContext(c.Request.Context())
	if tx == nil {
		c.JSON(500, gin.H{"error": "transaction not found"})
		return
	}

	// Simulate database query
	db := logmanager.StartDatabaseSegment(tx, logmanager.DatabaseSegment{
		Name:  "SELECT users",
		Table: "users",
		Query: "SELECT * FROM users",
		Host:  "localhost:5432",
	})
	defer db.End()

	// Simulate some processing
	time.Sleep(50 * time.Millisecond)

	c.JSON(200, gin.H{
		"users": []gin.H{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
			{"id": 3, "name": "Charlie"},
		},
	})
}

// getUser returns a single user with external API call simulation
func getUser(c *gin.Context) {
	tx := logmanager.FromContext(c.Request.Context())
	if tx == nil {
		c.JSON(500, gin.H{"error": "transaction not found"})
		return
	}

	id := c.Param("id")

	// Simulate database query
	db := logmanager.StartDatabaseSegment(tx, logmanager.DatabaseSegment{
		Name:  "SELECT user by ID",
		Table: "users",
		Query: fmt.Sprintf("SELECT * FROM users WHERE id = %s", id),
		Host:  "localhost:5432",
	})
	defer db.End()

	// Simulate processing
	time.Sleep(30 * time.Millisecond)

	c.JSON(200, gin.H{
		"id":   id,
		"name": "User " + id,
	})
}

// createOrder creates a new order with external service call simulation
func createOrder(c *gin.Context) {
	tx := logmanager.FromContext(c.Request.Context())
	if tx == nil {
		c.JSON(500, gin.H{"error": "transaction not found"})
		return
	}

	var order struct {
		UserID int `json:"user_id"`
	}

	if err := c.BindJSON(&order); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	// Simulate external API call to payment service
	paymentReq, _ := http.NewRequest("POST", "http://payment-service/charge", nil)
	apiSegment := logmanager.StartApiSegment(logmanager.ApiSegment{
		Name:    "POST /payment-service/charge",
		Request: paymentReq,
	})
	defer apiSegment.End()

	// Simulate external service latency
	time.Sleep(100 * time.Millisecond)

	// Simulate database insert
	db := logmanager.StartDatabaseSegment(tx, logmanager.DatabaseSegment{
		Name:  "INSERT order",
		Table: "orders",
		Query: "INSERT INTO orders (user_id, status) VALUES ($1, 'pending')",
		Host:  "localhost:5432",
	})
	defer db.End()

	// Simulate processing
	time.Sleep(40 * time.Millisecond)

	c.JSON(201, gin.H{
		"order_id": 12345,
		"user_id":  order.UserID,
		"status":   "pending",
		"message":  "Order created successfully",
	})
}
