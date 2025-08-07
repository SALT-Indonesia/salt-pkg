package logmanager_test

import (
	"context"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal/test/testdata"
	"github.com/stretchr/testify/assert"
)

// TestInfoLoggingWithContext tests the LogInfoWithContext function
// This test verifies that the function handles different input scenarios correctly
func TestInfoLoggingWithContext(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		msg          string
		fields       map[string]string
		shouldPanic  bool
	}{
		{
			name: "empty message should not panic",
			setupContext: func() context.Context {
				return context.Background()
			},
			msg:         "",
			fields:      nil,
			shouldPanic: false,
		},
		{
			name: "nil context should not panic",
			setupContext: func() context.Context {
				return nil
			},
			msg:         "test message",
			fields:      nil,
			shouldPanic: false,
		},
		{
			name: "context with transaction should not panic",
			setupContext: func() context.Context {
				txn := testdata.NewTx("trace123", "name")
				return logmanager.NewContext(context.Background(), txn)
			},
			msg:         "test message with transaction",
			fields:      nil,
			shouldPanic: false,
		},
		{
			name: "context with trace ID but no transaction should not panic",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), logmanager.TraceIDContextKey.String(), "trace456")
			},
			msg:         "test message with trace ID",
			fields:      nil,
			shouldPanic: false,
		},
		{
			name: "with optional fields should not panic",
			setupContext: func() context.Context {
				return context.Background()
			},
			msg: "test message with fields",
			fields: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			shouldPanic: false,
		},
		{
			name: "with empty fields map should not panic",
			setupContext: func() context.Context {
				return context.Background()
			},
			msg:         "test message with empty fields",
			fields:      map[string]string{},
			shouldPanic: false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup context
			ctx := tt.setupContext()

			// Define a function to call the function under test
			testFunc := func() {
				if tt.fields != nil {
					logmanager.LogInfoWithContext(ctx, tt.msg, tt.fields)
				} else {
					logmanager.LogInfoWithContext(ctx, tt.msg)
				}
			}

			// Check if the function panics
			if tt.shouldPanic {
				assert.Panics(t, testFunc)
			} else {
				assert.NotPanics(t, testFunc)
			}
		})
	}
}