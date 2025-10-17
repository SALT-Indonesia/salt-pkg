package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

const traceIDKey = "xid"

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required" mask:"password"`
}

type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	r := gin.Default()

	app := logmanager.NewApplication(
		logmanager.WithAppName("gin-resty-example"),
		logmanager.WithTraceIDContextKey(traceIDKey),
		logmanager.WithExposeHeaders("X-Forwarded-For", "X-Url-Payload"),
	)

	// Initialize Resty client
	restyClient := resty.New().
		SetTimeout(10 * time.Second).
		SetBaseURL("https://httpbin.org")

	r.Use(traceIDMiddleware(), lmgin.Middleware(app))

	// Route 1: Simple registration that calls external API
	r.POST("/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Call external API using Resty with trace context
		user, err := createUserViaAPI(c.Request.Context(), restyClient, req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "registration successful",
			"user":    user,
		})
	})

	// Route 2: Get user info from external API
	r.GET("/user/:id", func(c *gin.Context) {
		userID := c.Param("id")

		// Call external API using Resty with trace context
		user, err := getUserFromAPI(c.Request.Context(), restyClient, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"user": user,
		})
	})

	// Route 3: Proxy endpoint that demonstrates multiple API calls
	r.POST("/process", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Make multiple API calls
		result, err := processWithMultipleAPICalls(c.Request.Context(), restyClient, req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, result)
	})

	fmt.Println("Gin + Resty server running at http://localhost:8002")
	if err := r.Run(":8002"); err != nil {
		panic(err)
	}
}

func traceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(traceIDKey, uuid.NewString())
		c.Next()
	}
}

// createUserViaAPI demonstrates Resty client call with logmanager integration
func createUserViaAPI(ctx context.Context, client *resty.Client, req RegisterRequest) (*UserResponse, error) {
	// Start a segment for this operation
	txn := logmanager.StartOtherSegment(
		logmanager.FromContext(ctx),
		logmanager.OtherSegment{
			Name: "create-user-api-call",
		},
	)
	defer txn.End()

	// Prepare masking config for sensitive data
	maskingConfigs := []logmanager.MaskingConfig{
		{
			FieldPattern: "password",
			Type:         logmanager.FullMask,
		},
		{
			FieldPattern: "email",
			Type:         logmanager.PartialMask,
			ShowFirst:    3,
			ShowLast:     10,
		},
	}

	// Make API call with context to propagate trace ID
	resp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&UserResponse{}).
		Post("/post")

	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Log the Resty transaction with masking
	restyTxn := lmresty.NewTxnWithConfig(resp, maskingConfigs)
	if restyTxn != nil {
		defer restyTxn.End()
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Return mock user data
	return &UserResponse{
		ID:       123,
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

// getUserFromAPI demonstrates GET request with logmanager
func getUserFromAPI(ctx context.Context, client *resty.Client, userID string) (*UserResponse, error) {
	txn := logmanager.StartOtherSegment(
		logmanager.FromContext(ctx),
		logmanager.OtherSegment{
			Name: "get-user-api-call",
		},
	)
	defer txn.End()

	resp, err := client.R().
		SetContext(ctx).
		SetQueryParam("user_id", userID).
		Get("/get")

	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Log the Resty transaction
	restyTxn := lmresty.NewTxn(resp)
	if restyTxn != nil {
		defer restyTxn.End()
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	time.Sleep(50 * time.Millisecond)

	return &UserResponse{
		ID:       123,
		Username: "john_doe",
		Email:    "john@example.com",
	}, nil
}

// processWithMultipleAPICalls demonstrates multiple sequential API calls with trace propagation
func processWithMultipleAPICalls(ctx context.Context, client *resty.Client, req RegisterRequest) (map[string]interface{}, error) {
	txn := logmanager.StartOtherSegment(
		logmanager.FromContext(ctx),
		logmanager.OtherSegment{
			Name: "process-multiple-api-calls",
		},
	)
	defer txn.End()

	// First API call - Create user
	user, err := createUserViaAPI(ctx, client, req)
	if err != nil {
		return nil, err
	}

	// Second API call - Verify email (simulated)
	verifyResp, err := client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"email": req.Email,
		}).
		Post("/post")

	if err != nil {
		return nil, fmt.Errorf("email verification failed: %w", err)
	}

	// Log verification API call with email masking
	verifyTxn := lmresty.NewTxnWithEmailMasking(verifyResp)
	if verifyTxn != nil {
		verifyTxn.End()
	}

	// Third API call - Send welcome notification (simulated)
	notifyResp, err := client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"user_id": user.ID,
			"message": "Welcome to our platform!",
		}).
		Post("/post")

	if err != nil {
		return nil, fmt.Errorf("notification failed: %w", err)
	}

	// Log notification API call
	notifyTxn := lmresty.NewTxn(notifyResp)
	if notifyTxn != nil {
		notifyTxn.End()
	}

	time.Sleep(150 * time.Millisecond)

	return map[string]interface{}{
		"status":          "completed",
		"user":            user,
		"email_verified":  true,
		"notification_sent": true,
	}, nil
}
