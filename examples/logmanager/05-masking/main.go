package main

import (
	"context"
	"fmt"
	"log"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
	"github.com/go-resty/resty/v2"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email" mask:"email"`
	Password string `json:"password" mask:"password"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password" mask:"password"`
}

type CreditCardRequest struct {
	CardNumber string  `json:"card_number"`
	CVV        string  `json:"cvv"`
	Amount     float64 `json:"amount"`
}

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("masking-examples"),
		logmanager.WithDebug(),
	)

	client := resty.New()

	fmt.Println("Running masking examples...")

	demonstrateCustomMasking(app, client)
	demonstratePasswordMasking(app, client)
	demonstrateEmailMasking(app, client)
	demonstrateCreditCardMasking(app, client)

	fmt.Println("All masking examples completed")
}

func demonstrateCustomMasking(app *logmanager.Application, client *resty.Client) {
	fmt.Println("1. Custom masking configuration")

	txn := app.StartHttp("trace-001", "POST /users")
	ctx := txn.ToContext(context.Background())

	maskingConfigs := []logmanager.MaskingConfig{
		{
			// EmailMask preserves domain and masks username middle
			// Example: john.doe@example.com -> jo****oe@example.com
			FieldPattern: "email",
			Type:         logmanager.EmailMask,
			ShowFirst:    2, // Show first 2 chars of username
			ShowLast:     2, // Show last 2 chars of username
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

	txnResp := lmresty.NewTxnWithMasking(resp, maskingConfigs)
	if txnResp != nil {
		txnResp.End()
	}

	fmt.Println("✓ Custom masking completed")
}

func demonstratePasswordMasking(app *logmanager.Application, client *resty.Client) {
	fmt.Println("2. Password masking convenience function")

	txn := app.StartHttp("trace-002", "POST /auth")
	ctx := txn.ToContext(context.Background())

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

	txnResp := lmresty.NewTxnWithPasswordMasking(resp)
	if txnResp != nil {
		txnResp.End()
	}

	fmt.Println("✓ Password masking completed")
}

func demonstrateEmailMasking(app *logmanager.Application, client *resty.Client) {
	fmt.Println("3. Email masking convenience function")
	fmt.Println("   Uses EmailMask type: preserves domain, masks username middle")
	fmt.Println("   Example: arfan.azhari@salt.id -> ar********ri@salt.id")

	txn := app.StartHttp("trace-003", "POST /users")
	ctx := txn.ToContext(context.Background())

	resp, err := client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"username":     "johndoe",
			"email":        "arfan.azhari@salt.id",
			"backup_email": "backup.user@example.com",
			"work_email":   "a@test.com", // Single char username edge case
		}).
		Post("https://httpbin.org/post")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	// NewTxnWithEmailMasking uses EmailMask type with ShowFirst=2, ShowLast=2
	// Results:
	// - arfan.azhari@salt.id     -> ar********ri@salt.id
	// - backup.user@example.com  -> ba********er@example.com
	// - a@test.com               -> *@test.com (single char edge case)
	txnResp := lmresty.NewTxnWithEmailMasking(resp)
	if txnResp != nil {
		txnResp.End()
	}

	fmt.Println("✓ Email masking completed")
}

func demonstrateCreditCardMasking(app *logmanager.Application, client *resty.Client) {
	fmt.Println("4. Credit card masking convenience function")

	txn := app.StartHttp("trace-004", "POST /payment")
	ctx := txn.ToContext(context.Background())

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

	txnResp := lmresty.NewTxnWithCreditCardMasking(resp)
	if txnResp != nil {
		txnResp.End()
	}

	fmt.Println("✓ Credit card masking completed")
}