package logmanager_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxnRecord_SetConsumer(t *testing.T) {
	tests := []struct {
		name        string
		app         *logmanager.TestableApplication
		traceID     string
		txnName     string
		consumer    *logmanager.Consumer
		expectedNil bool
		expectLog   bool
	}{
		{
			name:    "valid transaction and consumer with comprehensive field assertions",
			app:     logmanager.NewTestableApplication(),
			traceID: "consumer-trace-id",
			txnName: "ProcessMessage",
			consumer: &logmanager.Consumer{
				Exchange:    "orders-exchange",
				Queue:       "orders-queue",
				RoutingKey:  "order.created",
				RequestBody: []byte(`{"orderId": "12345", "amount": 100.50}`),
			},
			expectedNil: false,
			expectLog:   true,
		},
		{
			name:        "valid transaction with nil consumer",
			app:         logmanager.NewTestableApplication(),
			traceID:     "nil-consumer-trace",
			txnName:     "ProcessEmptyMessage",
			consumer:    nil,
			expectedNil: false,
			expectLog:   true,
		},
		{
			name:    "valid transaction with empty consumer data",
			app:     logmanager.NewTestableApplication(),
			traceID: "empty-consumer-trace",
			txnName: "ProcessDefaultMessage",
			consumer: &logmanager.Consumer{
				Exchange:    "",
				Queue:       "",
				RoutingKey:  "",
				RequestBody: []byte(""),
			},
			expectedNil: false,
			expectLog:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logged entries before each test
			tt.app.ResetLoggedEntries()

			// Create transaction and TxnRecord using the same pattern as testdata.NewTxRecord
			tx := tt.app.Application.StartHttp(tt.traceID, tt.txnName)
			txnRecord := tx.AddDatabase("db") // This returns a TxnRecord that supports SetConsumer
			txnRecord.SetConsumer(tt.consumer)
			txnRecord.End()
			tx.End() // End the transaction to trigger logging

			if tt.expectedNil {
				assert.Nil(t, txnRecord)
				return
			}
			assert.NotNil(t, txnRecord)

			if tt.expectLog {
				// Assert logged data keys and values - expect 2 entries (database/consumer + HTTP transaction)
				assert.Equal(t, 2, tt.app.CountLoggedEntries(), "Should have exactly two logged entries")

				// Get all logged entries to verify both
				entries := tt.app.GetLoggedEntries()

				// First entry should be the database/consumer log
				firstEntry := entries[0]
				assert.Equal(t, tt.traceID, firstEntry.Data["trace_id"], "First entry should log correct trace_id")
				assert.Equal(t, logmanager.TxnTypeDatabase, firstEntry.Data["type"], "First entry should be database type")
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
