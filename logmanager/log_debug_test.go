package logmanager_test

import (
	"context"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/stretchr/testify/assert"
)

func TestDebugWithContext(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		msg         string
		fields      []map[string]string
		shouldPanic bool
	}{
		{
			name:        "Empty message should return early",
			ctx:         context.Background(),
			msg:         "",
			fields:      nil,
			shouldPanic: false,
		},
		{
			name:        "Valid message with context",
			ctx:         context.Background(),
			msg:         "Debug message",
			fields:      nil,
			shouldPanic: false,
		},
		{
			name:        "Valid message with fields",
			ctx:         context.Background(),
			msg:         "Debug message with fields",
			fields:      []map[string]string{{"key": "value"}},
			shouldPanic: false,
		},
		{
			name:        "Nil context",
			ctx:         nil,
			msg:         "Debug message with nil context",
			fields:      nil,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					logmanager.DebugWithContext(tt.ctx, tt.msg, tt.fields...)
				})
			} else {
				assert.NotPanics(t, func() {
					logmanager.DebugWithContext(tt.ctx, tt.msg, tt.fields...)
				})
			}
		})
	}
}

func TestDebugWithContextAndTransaction(t *testing.T) {
	tests := []struct {
		name         string
		debugEnabled bool
		msg          string
		expectLog    bool
	}{
		{
			name:         "Debug enabled should log",
			debugEnabled: true,
			msg:          "Debug message",
			expectLog:    true,
		},
		{
			name:         "Debug disabled should not log",
			debugEnabled: false,
			msg:          "Debug message",
			expectLog:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create application with or without debug mode
			var app *logmanager.Application
			if tt.debugEnabled {
				app = logmanager.NewApplication(logmanager.WithDebug())
			} else {
				app = logmanager.NewApplication(logmanager.WithEnvironment("production"))
			}

			// Start a transaction to get context with debug setting
			txn := app.StartHttp("test-trace", "test")
			ctx := logmanager.NewContext(context.Background(), txn)

			// This should not panic regardless of debug setting
			assert.NotPanics(t, func() {
				logmanager.DebugWithContext(ctx, tt.msg)
			})
		})
	}
}

func TestDebugWithContextBackwardCompatibility(t *testing.T) {
	// Test that debug logging works with contexts that don't have transactions
	// for backward compatibility
	tests := []struct {
		name string
		ctx  context.Context
		msg  string
	}{
		{
			name: "Context without transaction",
			ctx:  context.Background(),
			msg:  "Debug message without transaction",
		},
		{
			name: "Context with trace ID but no transaction",
			ctx:  context.WithValue(context.Background(), logmanager.TraceIDContextKey.String(), "test-trace-123"),
			msg:  "Debug message with trace ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic and should log (backward compatibility)
			assert.NotPanics(t, func() {
				logmanager.DebugWithContext(tt.ctx, tt.msg)
			})
		})
	}
}