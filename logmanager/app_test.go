package logmanager_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartConsumer(t *testing.T) {
	tests := []struct {
		name           string
		app            *logmanager.Application
		traceID        string
		expectedTxnNil bool
	}{
		{
			name:           "Nil application",
			app:            nil,
			traceID:        "trace123",
			expectedTxnNil: false,
		},
		{
			name:           "Valid application with trace and name",
			app:            logmanager.NewApplication(),
			traceID:        "trace123",
			expectedTxnNil: false,
		},
		{
			name:           "Valid application without trace",
			app:            logmanager.NewApplication(),
			traceID:        "",
			expectedTxnNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn := tt.app.StartConsumer(tt.traceID)
			assert.NotNil(t, txn)
			assert.Equal(t, tt.expectedTxnNil, txn == nil)
		})
	}
}

func TestTraceIDContextKey(t *testing.T) {
	tests := []struct {
		name               string
		app                *logmanager.Application
		expectedContextKey logmanager.ContextKey
	}{
		{
			name:               "Nil application",
			app:                nil,
			expectedContextKey: "",
		},
		{
			name:               "Application with default TraceIDContextKey",
			app:                logmanager.NewApplication(),
			expectedContextKey: logmanager.TraceIDContextKey,
		},
		{
			name:               "Application with TraceIDContextKey set",
			app:                logmanager.NewApplication(logmanager.WithTraceIDContextKey("X-Custom-Key")),
			expectedContextKey: "X-Custom-Key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.app.TraceIDContextKey()
			assert.Equal(t, tt.expectedContextKey, result)
		})
	}
}

func TestTraceIDHeaderKey(t *testing.T) {
	tests := []struct {
		name              string
		app               *logmanager.Application
		expectedHeaderKey string
	}{
		{
			name:              "Nil application",
			app:               nil,
			expectedHeaderKey: "",
		},
		{
			name:              "Application with TraceIDHeaderKey set",
			app:               logmanager.NewApplication(logmanager.WithTraceIDHeaderKey("X-Custom-ID")),
			expectedHeaderKey: "X-Custom-ID",
		},
		{
			name:              "Application with no TraceIDHeaderKey",
			app:               logmanager.NewApplication(),
			expectedHeaderKey: "X-Trace-Id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.app.TraceIDHeaderKey()
			assert.Equal(t, tt.expectedHeaderKey, result)
		})
	}
}

func TestTraceIDViaHeader(t *testing.T) {
	tests := []struct {
		name         string
		app          *logmanager.Application
		expectedBool bool
	}{
		{
			name:         "Nil application",
			app:          nil,
			expectedBool: false,
		},
		{
			name:         "Application with TraceIDViaHeader true",
			app:          logmanager.NewApplication(logmanager.WithTraceIDHeaderKey("X-Trace-ID")),
			expectedBool: true,
		},
		{
			name:         "Application with TraceIDViaHeader false",
			app:          logmanager.NewApplication(),
			expectedBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.app.TraceIDViaHeader()
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}

func TestNewApplication(t *testing.T) {
	tests := []struct {
		name          string
		options       []logmanager.Option
		expectedName  string
		expectedDebug bool
	}{
		{
			name:          "Default settings",
			options:       nil,
			expectedName:  "default",
			expectedDebug: false,
		},
		{
			name:          "Custom name",
			options:       []logmanager.Option{logmanager.WithAppName("customApp")},
			expectedName:  "customApp",
			expectedDebug: false,
		},
		{
			name:          "Debug mode enabled",
			options:       []logmanager.Option{logmanager.WithDebug()},
			expectedName:  "default",
			expectedDebug: true,
		},
		{
			name:          "Custom name and debug mode enabled",
			options:       []logmanager.Option{logmanager.WithAppName("customApp"), logmanager.WithDebug()},
			expectedName:  "customApp",
			expectedDebug: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication(tt.options...)
			assert.NotNil(t, app)
		})
	}
}

func TestStartHttp(t *testing.T) {
	tests := []struct {
		name           string
		app            *logmanager.Application
		traceID        string
		httpName       string
		expectedTxnNil bool
	}{
		{
			name:           "Nil Application",
			app:            nil,
			traceID:        "trace123",
			httpName:       "http1",
			expectedTxnNil: false,
		},
		{
			name:           "Valid Application with trace and name",
			app:            logmanager.NewApplication(),
			traceID:        "trace123",
			httpName:       "http1",
			expectedTxnNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn := tt.app.StartHttp(tt.traceID, tt.httpName)
			assert.NotNil(t, txn, "Transaction should not be nil")
			assert.Equal(t, tt.expectedTxnNil, txn == nil, "Equality check for expected nil transaction")
		})
	}
}

func TestStartOther(t *testing.T) {
	tests := []struct {
		name           string
		app            *logmanager.Application
		traceID        string
		httpName       string
		expectedTxnNil bool
	}{
		{
			name:           "Nil Application",
			app:            nil,
			traceID:        "trace123",
			httpName:       "http1",
			expectedTxnNil: false,
		},
		{
			name:           "Valid Application with trace and name",
			app:            logmanager.NewApplication(),
			traceID:        "trace123",
			httpName:       "http1",
			expectedTxnNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn := tt.app.Start(tt.traceID, tt.httpName, logmanager.TxnTypeOther)
			assert.NotNil(t, txn, "Transaction should not be nil")
			assert.Equal(t, tt.expectedTxnNil, txn == nil, "Equality check for expected nil transaction")
		})
	}
}
