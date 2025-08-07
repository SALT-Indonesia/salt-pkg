package logmanager

import (
	"context"
	"fmt"
)

// ExampleLogInfoWithContext demonstrates the usage of LogInfoWithContext function
func ExampleLogInfoWithContext() {
	// Example 1: Basic usage with context and message only
	ctx := context.Background()
	LogInfoWithContext(ctx, "Service started successfully")

	// Example 2: Usage with trace ID in context
	ctxWithTrace := context.WithValue(ctx, TraceIDContextKey.String(), "trace-12345")
	LogInfoWithContext(ctxWithTrace, "User authentication completed")

	// Example 3: Usage with optional fields
	fields := map[string]string{
		"user_id":    "user-789",
		"session_id": "session-abc",
		"action":     "login",
	}
	LogInfoWithContext(ctxWithTrace, "User logged in successfully", fields)

	// Example 4: Usage with transaction context
	app := NewApplication()
	txn := app.StartHttp("trace-example", "example-service")
	txnCtx := NewContext(ctx, txn)
	
	LogInfoWithContext(txnCtx, "Processing user request")
	
	// Example 5: Usage with empty fields (should work fine)
	emptyFields := map[string]string{}
	LogInfoWithContext(txnCtx, "Empty fields example", emptyFields)

	txn.End()

	fmt.Println("LogInfoWithContext examples completed")
}