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
			FieldPattern: "email",
			Type:         logmanager.PartialMask,
			ShowFirst:    3,
			ShowLast:     10,
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

	txn := app.StartHttp("trace-003", "POST /users")
	ctx := txn.ToContext(context.Background())

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