package logmanager_test

import (
	"errors"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxnRecord_MarkAsErrorBusiness(t *testing.T) {
	tests := []struct {
		name      string
		app       *logmanager.TestableApplication
		traceID   string
		txnName   string
		wantNil   bool
		expectLog bool
	}{
		{
			name:      "Valid transaction marked as business error",
			app:       logmanager.NewTestableApplication(),
			traceID:   "test-trace-id",
			txnName:   "POST /api/payment",
			wantNil:   false,
			expectLog: true,
		},
		{
			name:      "Valid transaction with business error and field assertions",
			app:       logmanager.NewTestableApplication(),
			traceID:   "business-error-trace",
			txnName:   "PUT /api/order",
			wantNil:   false,
			expectLog: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logged entries before each test
			tt.app.ResetLoggedEntries()

			tx := tt.app.Application.StartHttp(tt.traceID, tt.txnName)
			tx.MarkAsBusinessError()
			tx.End() // End the transaction to trigger logging

			if tt.wantNil {
				assert.Nil(t, tx.TxnRecord)
				return
			}
			assert.NotNil(t, tx.TxnRecord)

			if tt.expectLog {
				// Assert logged data keys and values
				assert.Equal(t, 1, tt.app.CountLoggedEntries(), "Should have exactly one logged entry")

				// Verify essential logged fields exist
				assert.True(t, tt.app.HasLoggedField("trace_id"), "Should log trace_id field")
				assert.True(t, tt.app.HasLoggedField("name"), "Should log name field")
				assert.True(t, tt.app.HasLoggedField("type"), "Should log type field")
				assert.True(t, tt.app.HasLoggedField("start"), "Should log start field")
				assert.True(t, tt.app.HasLoggedField("latency"), "Should log latency field")
				assert.True(t, tt.app.HasLoggedField("service"), "Should log service field")

				// Verify logged field values
				assert.Equal(t, tt.traceID, tt.app.GetLoggedField("trace_id"), "Should log correct trace_id")
				assert.Equal(t, tt.txnName, tt.app.GetLoggedField("name"), "Should log correct transaction name")
				assert.Equal(t, logmanager.TxnTypeHttp, tt.app.GetLoggedField("type"), "Should log HTTP transaction type")
				assert.Equal(t, "default", tt.app.GetLoggedField("service"), "Should log default service name")

				// Verify log level is Warning for business error transactions
				assert.Equal(t, logrus.WarnLevel, tt.app.GetLoggedLevel(), "Should log at Warning level for business error transactions")
			}
		})
	}
}

func TestTxnRecord_SetErrorBusiness(t *testing.T) {
	tests := []struct {
		name      string
		app       *logmanager.TestableApplication
		traceID   string
		txnName   string
		err       error
		expectNil bool
		expectLog bool
	}{
		{
			name:      "Set business error on valid transaction",
			app:       logmanager.NewTestableApplication(),
			traceID:   "error-trace-id",
			txnName:   "POST /api/payment",
			err:       errors.New("payment processing failed"),
			expectNil: false,
			expectLog: true,
		},
		{
			name:      "Set nil error on valid transaction",
			app:       logmanager.NewTestableApplication(),
			traceID:   "nil-error-trace",
			txnName:   "GET /api/status",
			err:       nil,
			expectNil: false,
			expectLog: true,
		},
		{
			name:      "Set business error with comprehensive field assertions",
			app:       logmanager.NewTestableApplication(),
			traceID:   "comprehensive-trace",
			txnName:   "PUT /api/user/profile",
			err:       errors.New("validation error: invalid email format"),
			expectNil: false,
			expectLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logged entries before each test
			tt.app.ResetLoggedEntries()

			tx := tt.app.Application.StartHttp(tt.traceID, tt.txnName)
			txnRecord := tx.AddDatabase("test-db") // Get TxnRecord to call SetBusinessError
			txnRecord.SetBusinessError(tt.err)
			tx.End() // End the transaction to trigger logging

			if tt.expectNil {
				assert.Nil(t, txnRecord)
				return
			}
			assert.NotNil(t, txnRecord)

			if tt.expectLog {
				// Assert logged data keys and values
				assert.Equal(t, 1, tt.app.CountLoggedEntries(), "Should have exactly one logged entry")

				// Verify essential logged fields exist
				assert.True(t, tt.app.HasLoggedField("trace_id"), "Should log trace_id field")
				assert.True(t, tt.app.HasLoggedField("name"), "Should log name field")
				assert.True(t, tt.app.HasLoggedField("type"), "Should log type field")
				assert.True(t, tt.app.HasLoggedField("start"), "Should log start field")
				assert.True(t, tt.app.HasLoggedField("latency"), "Should log latency field")
				assert.True(t, tt.app.HasLoggedField("service"), "Should log service field")

				// Verify logged field values
				assert.Equal(t, tt.traceID, tt.app.GetLoggedField("trace_id"), "Should log correct trace_id")
				assert.Equal(t, tt.txnName, tt.app.GetLoggedField("name"), "Should log correct transaction name")
				assert.Equal(t, logmanager.TxnTypeHttp, tt.app.GetLoggedField("type"), "Should log HTTP transaction type")
				assert.Equal(t, "default", tt.app.GetLoggedField("service"), "Should log default service name")

				// Verify log level is Error for all transactions (logmanager default behavior)
				assert.Equal(t, logrus.ErrorLevel, tt.app.GetLoggedLevel(), "Should log at Error level for transactions")
				assert.Equal(t, "internal server error", tt.app.GetLoggedMessage(), "Should have internal server error message")
			}
		})
	}
}

// Test End() method with comprehensive scenarios
func TestTxnRecord_End_BasicHTTP(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	tx := app.Application.StartHttp("http-trace-id", "GET /api/users")
	txn := tx.TxnRecord

	// End the transaction
	txn.End()

	// Verify transaction is not nil
	assert.NotNil(t, txn)

	// Assert logged entry count
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")

	// Verify essential logged fields exist
	assert.True(t, app.HasLoggedField("trace_id"), "Should log trace_id field")
	assert.True(t, app.HasLoggedField("name"), "Should log name field")
	assert.True(t, app.HasLoggedField("type"), "Should log type field")
	assert.True(t, app.HasLoggedField("start"), "Should log start field")
	assert.True(t, app.HasLoggedField("latency"), "Should log latency field")
	assert.True(t, app.HasLoggedField("service"), "Should log service field")

	// Verify logged field values
	assert.Equal(t, "http-trace-id", app.GetLoggedField("trace_id"), "Should log correct trace_id")
	assert.Equal(t, "GET /api/users", app.GetLoggedField("name"), "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeHttp, app.GetLoggedField("type"), "Should log HTTP transaction type")
	assert.Equal(t, "default", app.GetLoggedField("service"), "Should log default service name")

	// Verify log level is Error for basic transactions
	assert.Equal(t, logrus.ErrorLevel, app.GetLoggedLevel(), "Should log at Error level for transactions")
	assert.Equal(t, "internal server error", app.GetLoggedMessage(), "Should have internal server error message")
}

func TestTxnRecord_End_NilTransaction(t *testing.T) {
	var txn *logmanager.TxnRecord
	txn.End() // Should not panic
	assert.Nil(t, txn)
}

func TestTxnRecord_End_ZeroStartTime(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	txn := &logmanager.TxnRecord{} // Zero start time
	txn.End()

	// Should not log anything
	assert.Equal(t, 0, app.CountLoggedEntries(), "Should not log anything for zero start time")
}

func TestTxnRecord_End_MultipleEndCalls(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	tx := app.Application.StartHttp("multi-end-trace", "GET /test")
	txn := tx.TxnRecord

	// First End call should log
	txn.End()
	assert.Equal(t, 1, app.CountLoggedEntries(), "First End call should create log entry")

	// Reset for second test
	app.ResetLoggedEntries()

	// Second End call should not log (start time is reset)
	txn.End()
	assert.Equal(t, 0, app.CountLoggedEntries(), "Second End call should not create log entry")
}
