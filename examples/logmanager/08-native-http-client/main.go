package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Job   string `json:"job"`
}

type CreateUserResponse struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Job       string `json:"job"`
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
}

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("native-http-client-demo"),
	)

	txn := app.Start("native-http-examples", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	fmt.Println("=== Native HTTP Client Examples with Logmanager ===\n")

	// Example 1: Basic GET request
	makeBasicGetRequest(ctx)

	// Example 2: POST request with JSON body
	makePostRequest(ctx)

	// Example 3: PUT request
	makePutRequest(ctx)

	// Example 4: DELETE request
	makeDeleteRequest(ctx)

	// Example 5: Request with query parameters
	makeRequestWithQueryParams(ctx)

	// Example 6: Request with custom headers
	makeRequestWithHeaders(ctx)

	// Example 7: Request with timeout
	makeRequestWithTimeout(ctx)

	// Example 8: Request with custom client configuration
	makeRequestWithCustomClient(ctx)

	fmt.Println("\n=== All examples completed ===")
}

func makeBasicGetRequest(ctx context.Context) {
	fmt.Println("1. Basic GET Request")

	// Get transaction from context
	tx := logmanager.FromContext(ctx)
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/get", nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n\n", err)
		return
	}

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n\n", err)
		return
	}

	// Log the transaction
	txn := tx.AddTxnNow("GET-httpbin", logmanager.TxnTypeApi, startTime)
	txn.SetWebRequest(req)
	txn.SetResponseBodyAndCode(body, resp.StatusCode)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode)
	fmt.Printf("   ✓ Response size: %d bytes\n\n", len(body))
}

func makePostRequest(ctx context.Context) {
	fmt.Println("2. POST Request with JSON Body")

	tx := logmanager.FromContext(ctx)
	startTime := time.Now()

	// Prepare request body
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Job:   "Software Engineer",
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("   ❌ Error marshaling JSON: %v\n\n", err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://httpbin.org/post", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n\n", err)
		return
	}

	// Log the transaction
	txn := tx.AddTxnNow("POST-httpbin", logmanager.TxnTypeApi, startTime)
	txn.SetWebRequest(req)
	txn.SetResponseBodyAndCode(body, resp.StatusCode)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode)
	fmt.Printf("   ✓ User data sent: %s\n\n", user.Name)
}

func makePutRequest(ctx context.Context) {
	fmt.Println("3. PUT Request")

	tx := logmanager.FromContext(ctx)
	startTime := time.Now()

	// Prepare request body
	updateData := map[string]interface{}{
		"name": "Jane Doe",
		"job":  "Senior Developer",
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		fmt.Printf("   ❌ Error marshaling JSON: %v\n\n", err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "PUT", "https://httpbin.org/put", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n\n", err)
		return
	}

	// Log the transaction
	txn := tx.AddTxnNow("PUT-httpbin", logmanager.TxnTypeApi, startTime)
	txn.SetWebRequest(req)
	txn.SetResponseBodyAndCode(body, resp.StatusCode)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode)
	fmt.Printf("   ✓ Update completed successfully\n\n")
}

func makeDeleteRequest(ctx context.Context) {
	fmt.Println("4. DELETE Request")

	tx := logmanager.FromContext(ctx)
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "DELETE", "https://httpbin.org/delete", nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n\n", err)
		return
	}

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n\n", err)
		return
	}

	// Log the transaction
	txn := tx.AddTxnNow("DELETE-httpbin", logmanager.TxnTypeApi, startTime)
	txn.SetWebRequest(req)
	txn.SetResponseBodyAndCode(body, resp.StatusCode)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode)
	fmt.Printf("   ✓ Delete request completed\n\n")
}

func makeRequestWithQueryParams(ctx context.Context) {
	fmt.Println("5. Request with Query Parameters")

	tx := logmanager.FromContext(ctx)
	startTime := time.Now()

	// Create HTTP request with query parameters
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/get", nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n\n", err)
		return
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("page", "1")
	q.Add("limit", "10")
	q.Add("sort", "desc")
	q.Add("category", "technology")
	req.URL.RawQuery = q.Encode()

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n\n", err)
		return
	}

	// Log the transaction
	txn := tx.AddTxnNow("GET-with-params", logmanager.TxnTypeApi, startTime)
	txn.SetWebRequest(req)
	txn.SetResponseBodyAndCode(body, resp.StatusCode)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode)
	fmt.Printf("   ✓ Query parameters sent successfully\n\n")
}

func makeRequestWithHeaders(ctx context.Context) {
	fmt.Println("6. Request with Custom Headers")

	tx := logmanager.FromContext(ctx)
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/headers", nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n\n", err)
		return
	}

	// Add custom headers
	req.Header.Set("User-Agent", "Native-HTTP-Logmanager/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Request-ID", "req-native-123")
	req.Header.Set("X-Client-Version", "v1.0.0")
	req.Header.Set("X-Custom-Header", "custom-value")

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n\n", err)
		return
	}

	// Log the transaction
	txn := tx.AddTxnNow("GET-with-headers", logmanager.TxnTypeApi, startTime)
	txn.SetWebRequest(req)
	txn.SetResponseBodyAndCode(body, resp.StatusCode)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode)
	fmt.Printf("   ✓ Custom headers sent successfully\n\n")
}

func makeRequestWithTimeout(ctx context.Context) {
	fmt.Println("7. Request with Timeout")

	tx := logmanager.FromContext(ctx)
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/delay/1", nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n\n", err)
		return
	}

	// Create client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		// Log transaction with error
		txn := tx.AddTxnNow("GET-with-timeout", logmanager.TxnTypeApi, startTime)
		txn.SetWebRequest(req)
		txn.NoticeError(err)
		defer txn.End()

		fmt.Printf("   ❌ Error executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n\n", err)
		return
	}

	// Log the transaction
	txn := tx.AddTxnNow("GET-with-timeout", logmanager.TxnTypeApi, startTime)
	txn.SetWebRequest(req)
	txn.SetResponseBodyAndCode(body, resp.StatusCode)
	defer txn.End()

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode)
	fmt.Printf("   ✓ Request with timeout completed successfully\n\n")
}

func makeRequestWithCustomClient(ctx context.Context) {
	fmt.Println("8. Request with Custom Client Configuration")

	tx := logmanager.FromContext(ctx)
	startTime := time.Now()

	// Create custom HTTP client with transport configuration
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/get", nil)
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n\n", err)
		return
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Error executing request: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading response: %v\n\n", err)
		return
	}

	// Log the transaction
	txn := tx.AddTxnNow("GET-custom-client", logmanager.TxnTypeApi, startTime)
	txn.SetWebRequest(req)
	txn.SetResponseBodyAndCode(body, resp.StatusCode)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Printf("   ✓ Status: %d\n", resp.StatusCode)
	fmt.Printf("   ✓ Custom client configuration successful\n\n")
}
