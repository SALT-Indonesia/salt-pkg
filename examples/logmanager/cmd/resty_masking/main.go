package main

import (
	"context"
	"log"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
	"github.com/go-resty/resty/v2"
)

// User represents a user with sensitive data that should be masked
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email" mask:"email"`       // This will be masked with go-masker struct tags
	Password string `json:"password" mask:"password"` // This will be masked with go-masker struct tags
	Phone    string `json:"phone"`
}

// LoginRequest represents a login request with sensitive data
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password" mask:"password"` // Masked with struct tags
}

// CreditCardRequest represents a payment request with sensitive data
type CreditCardRequest struct {
	CardNumber string  `json:"card_number"`
	CVV        string  `json:"cvv"`
	Amount     float64 `json:"amount"`
}

func main() {
	// Initialize logmanager application
	app := logmanager.NewApplication(
		logmanager.WithService("resty-masking-example"),
		logmanager.WithDebug(),
	)

	// Create a new resty client
	client := resty.New()

	// Example 1: Using NewTxnWithMasking with custom configurations
	demonstrateCustomMasking(app, client)

	// Example 2: Using NewTxnWithConfig (recommended approach)
	demonstrateConfigBasedMasking(app, client)

	// Example 3: Using convenience functions
	demonstrateConvenienceFunctions(app, client)
}

func demonstrateCustomMasking(app *logmanager.Application, client *resty.Client) {
	// Start a new HTTP transaction
	txnFromApp := app.StartHttp("trace-001", "POST /users")
	ctx := txnFromApp.ToContext(context.Background())

	// Define custom masking configurations
	maskingConfigs := []logmanager.MaskingConfig{
		{
			FieldPattern: "email",
			Type:         logmanager.PartialMask,
			ShowFirst:    3,
			ShowLast:     10, // Show @domain.com part
		},
		{
			FieldPattern: "password",
			Type:         logmanager.FullMask,
		},
		{
			JSONPath:  "$.phone",
			Type:      logmanager.PartialMask,
			ShowFirst: 3,
			ShowLast:  4,
		},
	}

	// Make a request with sensitive data
	resp, err := client.R().
		SetContext(ctx).
		SetBody(User{
			ID:       1,
			Username: "johndoe",
			Email:    "john.doe@example.com",
			Password: "supersecret123",
			Phone:    "+1234567890",
		}).
		Post("https://httpbin.org/post")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	// Create a transaction record with masking
	txn := lmresty.NewTxnWithMasking(resp, maskingConfigs)
	if txn != nil {
		txn.End()
	}

	log.Println("✓ Custom masking demonstration completed")
}

func demonstrateConfigBasedMasking(app *logmanager.Application, client *resty.Client) {
	// Start a new HTTP transaction
	txnFromApp := app.StartHttp("trace-002", "POST /login")
	ctx := txnFromApp.ToContext(context.Background())

	// Define masking configurations for login scenario
	maskingConfigs := []logmanager.MaskingConfig{
		{
			FieldPattern: "password",
			Type:         logmanager.FullMask,
		},
		{
			FieldPattern: "token",
			Type:         logmanager.FullMask,
		},
		{
			FieldPattern: "secret",
			Type:         logmanager.FullMask,
		},
		{
			JSONPath: "$.authorization",
			Type:     logmanager.FullMask,
		},
	}

	// Make login request
	resp, err := client.R().
		SetContext(ctx).
		SetBody(LoginRequest{
			Username: "johndoe",
			Password: "mypassword123",
		}).
		Post("https://httpbin.org/post")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	// Create a transaction record using NewTxnWithConfig (recommended)
	txn := lmresty.NewTxnWithConfig(resp, maskingConfigs)
	if txn != nil {
		txn.End()
	}

	log.Println("✓ Config-based masking demonstration completed")
}

func demonstrateConvenienceFunctions(app *logmanager.Application, client *resty.Client) {
	// Example 1: Password masking
	demonstratePasswordMasking(app, client)

	// Example 2: Email masking
	demonstrateEmailMasking(app, client)

	// Example 3: Credit card masking
	demonstrateCreditCardMasking(app, client)
}

func demonstratePasswordMasking(app *logmanager.Application, client *resty.Client) {
	txnFromApp := app.StartHttp("trace-003", "POST /auth")
	ctx := txnFromApp.ToContext(context.Background())

	resp, err := client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"username":      "johndoe",
			"password":      "secret123",
			"client_secret": "oauth_secret",
			"token":         "bearer_token",
		}).
		Post("https://httpbin.org/post")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	// Use convenience function for password masking
	txn := lmresty.NewTxnWithPasswordMasking(resp)
	if txn != nil {
		txn.End()
	}

	log.Println("✓ Password masking convenience function demonstrated")
}

func demonstrateEmailMasking(app *logmanager.Application, client *resty.Client) {
	txnFromApp := app.StartHttp("trace-004", "POST /users")
	ctx := txnFromApp.ToContext(context.Background())

	resp, err := client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"username":     "johndoe",
			"email":        "john.doe@example.com",
			"backup_email": "backup@example.com",
		}).
		Post("https://httpbin.org/post")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	// Use convenience function for email masking
	txn := lmresty.NewTxnWithEmailMasking(resp)
	if txn != nil {
		txn.End()
	}

	log.Println("✓ Email masking convenience function demonstrated")
}

func demonstrateCreditCardMasking(app *logmanager.Application, client *resty.Client) {
	txnFromApp := app.StartHttp("trace-005", "POST /payment")
	ctx := txnFromApp.ToContext(context.Background())

	resp, err := client.R().
		SetContext(ctx).
		SetBody(CreditCardRequest{
			CardNumber: "4532123456789012",
			CVV:        "123",
			Amount:     99.99,
		}).
		Post("https://httpbin.org/post")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	// Use convenience function for credit card masking
	txn := lmresty.NewTxnWithCreditCardMasking(resp)
	if txn != nil {
		txn.End()
	}

	log.Println("✓ Credit card masking convenience function demonstrated")
}
