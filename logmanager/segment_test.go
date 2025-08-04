package logmanager_test

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal/test/testdata"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStartGrpcSegment(t *testing.T) {
	tests := []struct {
		name   string
		tx     *logmanager.Transaction
		i      logmanager.GrpcSegment
		wanNil bool
	}{
		{
			name: "it should return nil if ctx is nil",
			tx:   nil,
			i: logmanager.GrpcSegment{
				Url:     "/product.ProductService/GetProduct",
				Request: testdata.NewRandomData(),
			},
			wanNil: true,
		},
		{
			name: "it should be ok",
			tx:   testdata.NewTx("id", "name"),
			i: logmanager.GrpcSegment{
				Url:     "/product.ProductService/GetProduct",
				Request: testdata.NewRandomData(),
			},
			wanNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := logmanager.StartGrpcSegment(tt.tx, tt.i)
			if tt.wanNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
		})
	}
}

func TestStartOtherSegment(t *testing.T) {
	tests := []struct {
		name      string
		app       *logmanager.TestableApplication
		traceID   string
		txnName   string
		segment   logmanager.OtherSegment
		wantNil   bool
		expectLog bool
	}{
		{
			name:    "Valid transaction with other segment and comprehensive field assertions",
			app:     logmanager.NewTestableApplication(),
			traceID: "segment-trace-id",
			txnName: "GET /api/process",
			segment: logmanager.OtherSegment{
				Name:  "ProcessingSegment",
				Extra: map[string]interface{}{"key1": "value1"},
			},
			wantNil:   false,
			expectLog: true,
		},
		{
			name:    "Valid transaction with empty segment name",
			app:     logmanager.NewTestableApplication(),
			traceID: "empty-name-trace",
			txnName: "POST /api/calculate",
			segment: logmanager.OtherSegment{
				Name:  "",
				Extra: map[string]interface{}{"key1": "value1"},
			},
			wantNil:   false,
			expectLog: true,
		},
		{
			name:    "Valid transaction with nil Extra map",
			app:     logmanager.NewTestableApplication(),
			traceID: "nil-extra-trace",
			txnName: "PUT /api/update",
			segment: logmanager.OtherSegment{
				Name:  "UpdateSegment",
				Extra: nil,
			},
			wantNil:   false,
			expectLog: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logged entries before each test
			tt.app.ResetLoggedEntries()

			tx := tt.app.Application.StartHttp(tt.traceID, tt.txnName)
			got := logmanager.StartOtherSegment(tx, tt.segment)
			got.End()
			tx.End() // End the transaction to trigger logging

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)

			if tt.expectLog {
				// Assert logged data keys and values - expect 2 entries (segment + HTTP transaction)
				assert.Equal(t, 2, tt.app.CountLoggedEntries(), "Should have exactly two logged entries")

				// Get all logged entries to verify both
				entries := tt.app.GetLoggedEntries()

				// First entry should be the segment log
				firstEntry := entries[0]
				assert.Equal(t, tt.traceID, firstEntry.Data["trace_id"], "First entry should log correct trace_id")
				assert.Equal(t, logmanager.TxnTypeOther, firstEntry.Data["type"], "First entry should be other type")
				assert.Equal(t, logrus.InfoLevel, firstEntry.Level, "First entry should be Info level")
				assert.Equal(t, "", firstEntry.Message, "First entry should have empty message")

				// Second entry should be the HTTP transaction log
				secondEntry := entries[1]
				assert.Equal(t, tt.traceID, secondEntry.Data["trace_id"], "Second entry should log correct trace_id")
				assert.Equal(t, tt.txnName, secondEntry.Data["name"], "Second entry should log correct transaction name")
				assert.Equal(t, logmanager.TxnTypeHttp, secondEntry.Data["type"], "Second entry should be HTTP transaction type")
				assert.Equal(t, "default", secondEntry.Data["service"], "Second entry should log default service name")
				assert.Equal(t, logrus.ErrorLevel, secondEntry.Level, "Second entry should be Error level")
				assert.Equal(t, "internal server error", secondEntry.Message, "Second entry should have internal server error message")
			}
		})
	}
}

func TestStartOtherSegmentWithContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		i       logmanager.OtherSegment
		wantNil bool
	}{
		{
			name:    "it should return nil if context is nil",
			ctx:     nil,
			i:       logmanager.OtherSegment{Name: "TestSegment", Extra: map[string]interface{}{"key1": "value1"}},
			wantNil: true,
		},
		{
			name:    "it should return nil if transaction is not in context",
			ctx:     context.Background(),
			i:       logmanager.OtherSegment{Name: "TestSegment", Extra: map[string]interface{}{"key1": "value1"}},
			wantNil: true,
		},
		{
			name:    "it should generate name if name is empty",
			ctx:     testdata.NewTx("id", "name").ToContext(context.Background()),
			i:       logmanager.OtherSegment{Name: "", Extra: map[string]interface{}{"key1": "value1"}},
			wantNil: false,
		},
		{
			name:    "it should be ok with valid context and name",
			ctx:     testdata.NewTx("id", "name").ToContext(context.Background()),
			i:       logmanager.OtherSegment{Name: "TestSegment", Extra: map[string]interface{}{"key1": "value1"}},
			wantNil: false,
		},
		{
			name:    "it should be ok with empty Extra map",
			ctx:     testdata.NewTx("id", "name").ToContext(context.Background()),
			i:       logmanager.OtherSegment{Name: "TestSegment", Extra: map[string]interface{}{}},
			wantNil: false,
		},
		{
			name:    "it should be ok with nil Extra map",
			ctx:     testdata.NewTx("id", "name").ToContext(context.Background()),
			i:       logmanager.OtherSegment{Name: "TestSegment", Extra: nil},
			wantNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := logmanager.StartOtherSegmentWithContext(tt.ctx, tt.i)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
		})
	}
}

func TestStartOtherSegmentWithMessage(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		message string
	}{
		{
			name:    "it should not panic if context is nil",
			ctx:     nil,
			message: "test message",
		},
		{
			name:    "it should not panic if transaction is not in context",
			ctx:     context.Background(),
			message: "test message",
		},
		{
			name:    "it should be ok with valid context and message",
			ctx:     testdata.NewTx("id", "name").ToContext(context.Background()),
			message: "test message",
		},
		{
			name:    "it should be ok with empty message",
			ctx:     testdata.NewTx("id", "name").ToContext(context.Background()),
			message: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logmanager.StartOtherSegmentWithMessage(tt.ctx, tt.message)
		})
	}
}

func TestSetResponseValue(t *testing.T) {
	tests := []struct {
		name   string
		txn    *logmanager.TxnRecord
		value  interface{}
		wanNil bool
	}{
		{
			name:   "it should do nothing if txn is nil",
			txn:    nil,
			value:  "response data",
			wanNil: true,
		},
		{
			name:   "it should set the value successfully when txn and attrs are not nil",
			txn:    testdata.NewTx("id", "name").AddTxn("sub", logmanager.TxnTypeHttp),
			value:  "response data",
			wanNil: false,
		},
		{
			name:   "it should handle nil value gracefully",
			txn:    testdata.NewTx("id", "name").AddTxn("sub", logmanager.TxnTypeHttp),
			value:  nil,
			wanNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.txn.SetResponseValue(tt.value)
			if tt.wanNil {
				assert.Nil(t, tt.txn)
				return
			}
			assert.NotNil(t, tt.txn)
		})
	}
}

func TestStartApiSegment(t *testing.T) {
	tests := []struct {
		name               string
		i                  logmanager.ApiSegment
		wanNil             bool
		checkTraceIdHeader bool
		expectedTraceId    string
	}{
		{
			name: "it should be ok",
			i: logmanager.ApiSegment{
				Name:    "a",
				Request: testdata.NewRequestWithCtx(),
			},
			wanNil:             false,
			checkTraceIdHeader: true,
			expectedTraceId:    "1234567890", // This is the trace ID set in testdata.NewRequestWithCtx()
		},
		{
			name: "it should be nil with empty ctx",
			i: logmanager.ApiSegment{
				Name:    "a",
				Request: testdata.NewRequestWithEmptyCtx(),
			},
			wanNil:             true,
			checkTraceIdHeader: false,
		},
		{
			name: "it should be ok with empty name",
			i: logmanager.ApiSegment{
				Name:    "",
				Request: testdata.NewRequestWithCtx(),
			},
			wanNil:             false,
			checkTraceIdHeader: true,
			expectedTraceId:    "1234567890", // This is the trace ID set in testdata.NewRequestWithCtx()
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := customMethodName(tt.i)
			if tt.wanNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)

			// Check if X-Trace-Id header is set correctly
			if tt.checkTraceIdHeader {
				traceId := tt.i.Request.Header.Get("X-Trace-Id")
				assert.Equal(t, tt.expectedTraceId, traceId, "X-Trace-Id header should be set to the transaction's trace ID")
			}
		})
	}
}

func customMethodName(i logmanager.ApiSegment) *logmanager.TxnRecord {
	return logmanager.StartApiSegment(i)
}

func TestStartDatabaseSegment(t *testing.T) {
	tests := []struct {
		name    string
		tx      *logmanager.Transaction
		i       logmanager.DatabaseSegment
		wantNil bool
	}{
		{
			name: "it should be ok",
			tx:   testdata.NewTx("id", "name"),
			i: logmanager.DatabaseSegment{
				Name:  "repositoryProduct",
				Query: "select * from product",
				Table: "product",
				Host:  "localhost",
			},
			wantNil: false,
		},
		{
			name: "it should be generate name with empty name",
			tx:   testdata.NewTx("id", "name"),
			i: logmanager.DatabaseSegment{
				Query: "select * from product",
			},
			wantNil: false,
		},
		{
			name:    "it should be nil",
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := logmanager.StartDatabaseSegment(tt.tx, tt.i)
			got.End()
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
		})
	}
}
