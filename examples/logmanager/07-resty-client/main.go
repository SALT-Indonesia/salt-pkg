package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
	"github.com/go-resty/resty/v2"
)

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PostResponse struct {
	Data    interface{} `json:"data"`
	Headers interface{} `json:"headers"`
	URL     string      `json:"url"`
}

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("resty-client-demo"),
	)

	txn := app.Start("resty-examples", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	fmt.Println("=== Resty Client Examples with Logmanager ===\n")

	// Example 1: Basic GET request
	makeBasicGetRequest(ctx)

	// Example 2: POST request with JSON body
	makePostRequest(ctx)

	// Example 3: PUT request with headers
	makePutRequest(ctx)

	// Example 4: DELETE request
	makeDeleteRequest(ctx)

	// Example 5: Request with query parameters
	makeRequestWithQueryParams(ctx)

	// Example 6: Request with custom headers
	makeRequestWithHeaders(ctx)

	// Example 7: Request with masking (sensitive data)
	makeRequestWithMasking(ctx)

	// Example 8: Request with authentication
	makeAuthenticatedRequest(ctx)

	// Example 9: Request with timeout handling
	makeRequestWithTimeout(ctx)

	fmt.Println("\n=== All examples completed ===")
}

func makeBasicGetRequest(ctx context.Context) {
	fmt.Println("1. Basic GET Request")
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		Get("https://httpbin.org/get")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())
	fmt.Printf("   ✓ Response size: %d bytes\n\n", len(resp.Body()))
}

func makePostRequest(ctx context.Context) {
	fmt.Println("2. POST Request with JSON Body")
	client := resty.New()

	payload := map[string]interface{}{
		"title":  "Learning Resty with Logmanager",
		"body":   "This is a test post",
		"userId": 1,
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post("https://httpbin.org/post")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())
	fmt.Printf("   ✓ Payload sent: %v\n\n", payload)
}

func makePutRequest(ctx context.Context) {
	fmt.Println("3. PUT Request with Headers")
	client := resty.New()

	updateData := map[string]interface{}{
		"title": "Updated Title",
		"body":  "Updated content",
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Custom-Header", "custom-value").
		SetBody(updateData).
		Put("https://httpbin.org/put")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())
	fmt.Printf("   ✓ Update data sent successfully\n\n")
}

func makeDeleteRequest(ctx context.Context) {
	fmt.Println("4. DELETE Request")
	client := resty.New()

	resp, err := client.R().
		SetContext(ctx).
		Delete("https://httpbin.org/delete")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())
	fmt.Printf("   ✓ Delete request completed\n\n")
}

func makeRequestWithQueryParams(ctx context.Context) {
	fmt.Println("5. Request with Query Parameters")
	client := resty.New()

	resp, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"page":     "1",
			"limit":    "10",
			"sort":     "desc",
			"category": "technology",
		}).
		Get("https://httpbin.org/get")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())
	fmt.Printf("   ✓ Query parameters sent successfully\n\n")
}

func makeRequestWithHeaders(ctx context.Context) {
	fmt.Println("6. Request with Custom Headers")
	client := resty.New()

	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"User-Agent":      "Resty-Logmanager-Example/1.0",
			"Accept":          "application/json",
			"X-Request-ID":    "req-12345",
			"X-Client-Version": "v1.0.0",
		}).
		Get("https://httpbin.org/headers")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())
	fmt.Printf("   ✓ Custom headers sent successfully\n\n")
}

func makeRequestWithMasking(ctx context.Context) {
	fmt.Println("7. Request with Data Masking (Sensitive Data)")
	client := resty.New()

	user := User{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "SuperSecret123!",
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(user).
		Post("https://httpbin.org/post")

	// Use password masking to hide sensitive fields
	txn := lmresty.NewTxnWithPasswordMasking(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())
	fmt.Printf("   ✓ Sensitive data masked in logs\n\n")
}

func makeAuthenticatedRequest(ctx context.Context) {
	fmt.Println("8. Request with Authentication")
	client := resty.New()

	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken("fake-jwt-token-for-demo").
		SetHeader("Content-Type", "application/json").
		Get("https://httpbin.org/bearer")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err == nil {
		if authenticated, ok := result["authenticated"].(bool); ok && authenticated {
			fmt.Printf("   ✓ Authentication successful\n\n")
		} else {
			fmt.Printf("   ✓ Authentication token sent (validation may vary)\n\n")
		}
	}
}

func makeRequestWithTimeout(ctx context.Context) {
	fmt.Println("9. Request with Timeout Handling")
	client := resty.New()

	resp, err := client.R().
		SetContext(ctx).
		Get("https://httpbin.org/delay/1")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		fmt.Printf("   ❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode())
	fmt.Printf("   ✓ Request with delay completed successfully\n\n")
}
