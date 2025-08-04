package logmanager_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal/test/testdata"
	"github.com/stretchr/testify/assert"
)

// TestErrorLoggingWithContext tests the LogErrorWithContext function
// This test verifies that the function handles different input scenarios correctly
func TestErrorLoggingWithContext(t *testing.T) {
	// Create test cases
	tests := []struct {
		name         string
		setupContext func() context.Context
		err          error
		shouldPanic  bool
	}{
		{
			name: "nil error should not panic",
			setupContext: func() context.Context {
				return context.Background()
			},
			err:         nil,
			shouldPanic: false,
		},
		{
			name: "nil context should not panic",
			setupContext: func() context.Context {
				return nil
			},
			err:         errors.New("test error"),
			shouldPanic: false,
		},
		{
			name: "context with transaction should not panic",
			setupContext: func() context.Context {
				txn := testdata.NewTx("trace123", "name")
				return logmanager.NewContext(context.Background(), txn)
			},
			err:         errors.New("test error with transaction"),
			shouldPanic: false,
		},
		{
			name: "context with trace ID but no transaction should not panic",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), logmanager.TraceIDContextKey.String(), "trace456")
			},
			err:         errors.New("test error with trace ID"),
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
				logmanager.LogErrorWithContext(ctx, tt.err)
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
