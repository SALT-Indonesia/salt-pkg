package logmanager_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal/test/testdata"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTransaction_AddDatabase(t *testing.T) {
	tests := []struct {
		name      string
		app       *logmanager.TestableApplication
		traceID   string
		txnName   string
		inputName string
		expectNil bool
		expectLog bool
	}{
		{
			name:      "valid transaction with unique database name and field assertions",
			app:       logmanager.NewTestableApplication(),
			traceID:   "test-trace-db",
			txnName:   "GET /api/users",
			inputName: "users_db",
			expectNil: false,
			expectLog: true,
		},
		{
			name:      "valid transaction with another database name",
			app:       logmanager.NewTestableApplication(),
			traceID:   "test-trace-db2",
			txnName:   "POST /api/orders",
			inputName: "orders_db",
			expectNil: false,
			expectLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logged entries before each test
			tt.app.ResetLoggedEntries()

			transaction := tt.app.Application.StartHttp(tt.traceID, tt.txnName)
			record := transaction.AddDatabase(tt.inputName)
			transaction.End() // End the transaction to trigger logging

			if tt.expectNil {
				assert.Nil(t, record)
				return
			}
			assert.NotNil(t, record)

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

				// Verify log level is Error for transactions (logmanager default behavior)
				assert.Equal(t, logrus.ErrorLevel, tt.app.GetLoggedLevel(), "Should log at Error level for transactions")
				assert.Equal(t, "internal server error", tt.app.GetLoggedMessage(), "Should have internal server error message")
			}
		})
	}
}

func TestTransaction_AddTxn(t *testing.T) {
	tests := []struct {
		name        string
		transaction *logmanager.Transaction
		inputName   string
		inputType   logmanager.TxnType
		expectNil   bool
	}{
		{
			name: "valid transaction with valid txn type",
			transaction: testdata.NewTx(
				"test-trace",
				"test-service",
			),
			inputName: "transaction1",
			inputType: "type1",
			expectNil: false,
		},
		{
			name:        "nil transaction",
			transaction: nil,
			inputName:   "transaction2",
			inputType:   "type2",
			expectNil:   true,
		},
		{
			name: "valid transaction with duplicate transaction name",
			transaction: func() *logmanager.Transaction {
				tx := testdata.NewTx("test-trace", "test-service")
				tx.AddTxn("transaction1", "type1")
				return tx
			}(),
			inputName: "transaction1",
			inputType: "type1",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := tt.transaction.AddTxn(tt.inputName, tt.inputType)
			if tt.expectNil {
				assert.Nil(t, record)
			} else {
				assert.NotNil(t, record)
			}
		})
	}
}

func TestTransaction_AddTxnNow(t *testing.T) {
	tests := []struct {
		name        string
		transaction *logmanager.Transaction
		inputName   string
		inputType   logmanager.TxnType
		inputStart  time.Time
		expectedNil bool
	}{
		{
			name: "valid transaction with valid inputs",
			transaction: testdata.NewTx(
				"test-trace",
				"test-service",
			),
			inputName:   "transaction1",
			inputType:   "type1",
			inputStart:  time.Now(),
			expectedNil: false,
		},
		{
			name:        "nil transaction",
			transaction: nil,
			inputName:   "transaction2",
			inputType:   "type2",
			inputStart:  time.Now(),
			expectedNil: true,
		},
		{
			name: "duplicate transaction name",
			transaction: func() *logmanager.Transaction {
				tx := testdata.NewTx("test-trace", "test-service")
				tx.AddTxnNow("transaction1", "type1", time.Now())
				return tx
			}(),
			inputName:   "transaction1",
			inputType:   "type1",
			inputStart:  time.Now(),
			expectedNil: false,
		},
		{
			name: "empty transaction name",
			transaction: testdata.NewTx(
				"test-trace",
				"test-service",
			),
			inputName:   "",
			inputType:   "type1",
			inputStart:  time.Now(),
			expectedNil: false,
		},
		{
			name: "empty transaction type",
			transaction: testdata.NewTx(
				"test-trace",
				"test-service",
			),
			inputName:   "transaction3",
			inputType:   "",
			inputStart:  time.Now(),
			expectedNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := tt.transaction.AddTxnNow(tt.inputName, tt.inputType, tt.inputStart)
			if tt.expectedNil {
				assert.Nil(t, record)
			} else {
				assert.NotNil(t, record)
			}
		})
	}
}
