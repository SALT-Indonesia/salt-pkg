package logmanager_test

import (
	"errors"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxnRecord_NoticeError(t *testing.T) {
	tests := []struct {
		name string
		tx   *logmanager.TxnRecord
		err  error
	}{
		{
			name: "Notice error with nil transaction",
			tx:   nil,
			err:  nil,
		},
		{
			name: "Set error Successfully",
			tx:   &logmanager.TxnRecord{},
			err:  errors.New("test error"),
		},
		{
			name: "Reset error Successfully",
			tx:   &logmanager.TxnRecord{},
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tx.NoticeError(tt.err)
			if tt.tx == nil {
				assert.Nil(t, tt.tx)
				assert.Nil(t, tt.err)
				return
			}
			assert.NotNil(t, tt.tx)
		})
	}
}

func TestTxnRecord_End_APIWithBusinessError(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	parentTx := app.Application.StartHttp("api-business-error-trace", "parent")
	txn := parentTx.AddTxn("POST /api/payment", logmanager.TxnTypeApi)

	// Setup transaction with business error
	txn.SetBusinessError(errors.New("insufficient funds"))
	txn.SetResponseBodyAndCode([]byte(`{"error": "insufficient funds"}`), 400)

	// End the transaction
	txn.End()

	// Assert logged entry count
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")

	// Verify essential logged fields exist and values
	assert.True(t, app.HasLoggedField("trace_id"), "Should log trace_id field")
	assert.True(t, app.HasLoggedField("name"), "Should log name field")
	assert.True(t, app.HasLoggedField("type"), "Should log type field")
	assert.Equal(t, "api-business-error-trace", app.GetLoggedField("trace_id"), "Should log correct trace_id")
	assert.Equal(t, "POST /api/payment", app.GetLoggedField("name"), "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeApi, app.GetLoggedField("type"), "Should log API transaction type")

	// Check for response fields
	assert.True(t, app.HasLoggedField("response"), "Should log response body field")
	assert.True(t, app.HasLoggedField("status"), "Should log status code field")
	assert.Equal(t, 400, app.GetLoggedField("status"), "Should log correct status code")

	// Verify log level is Warning for business error transactions
	assert.Equal(t, logrus.WarnLevel, app.GetLoggedLevel(), "Should log at Warning level for business error transactions")
	assert.Equal(t, "insufficient funds", app.GetLoggedMessage(), "Should have business error message")
}

func TestTxnRecord_End_DatabaseWithError(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	parentTx := app.Application.StartHttp("db-error-trace", "parent")
	txn := parentTx.AddDatabase("SELECT users")

	// Setup transaction with error
	txn.NoticeError(errors.New("connection timeout"))

	// End the transaction
	txn.End()

	// Assert logged entry count
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")

	// Verify essential logged fields exist and values
	assert.True(t, app.HasLoggedField("trace_id"), "Should log trace_id field")
	assert.True(t, app.HasLoggedField("name"), "Should log name field")
	assert.True(t, app.HasLoggedField("type"), "Should log type field")
	assert.Equal(t, "db-error-trace", app.GetLoggedField("trace_id"), "Should log correct trace_id")
	assert.Equal(t, "SELECT users", app.GetLoggedField("name"), "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeDatabase, app.GetLoggedField("type"), "Should log Database transaction type")

	// Verify log level is Error for error transactions
	assert.Equal(t, logrus.ErrorLevel, app.GetLoggedLevel(), "Should log at Error level for error transactions")
	assert.Equal(t, "connection timeout", app.GetLoggedMessage(), "Should have error message")
}

func TestTxnRecord_End_HTTPSuccessful(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	parentTx := app.Application.StartHttp("http-success-trace", "parent")

	// Setup successful transaction
	parentTx.SetResponseBodyAndCode([]byte(`{"status": "success"}`), 200)

	// End the transaction
	parentTx.End()

	// Assert logged entry count
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")

	// Verify essential logged fields exist and values
	assert.True(t, app.HasLoggedField("trace_id"), "Should log trace_id field")
	assert.True(t, app.HasLoggedField("name"), "Should log name field")
	assert.True(t, app.HasLoggedField("type"), "Should log type field")
	assert.Equal(t, "http-success-trace", app.GetLoggedField("trace_id"), "Should log correct trace_id")
	assert.Equal(t, "parent", app.GetLoggedField("name"), "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeHttp, app.GetLoggedField("type"), "Should log HTTP transaction type")

	// Check for response fields
	assert.True(t, app.HasLoggedField("response"), "Should log response body field")
	assert.True(t, app.HasLoggedField("status"), "Should log status code field")
	assert.Equal(t, 200, app.GetLoggedField("status"), "Should log correct status code")

	// Verify log level is Info for successful transactions
	assert.Equal(t, logrus.InfoLevel, app.GetLoggedLevel(), "Should log at Info level for successful transactions")
	assert.Equal(t, "", app.GetLoggedMessage(), "Should have empty message for successful transactions")
}

func TestTxnRecord_End_ConsumerTransaction(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	// Start a consumer transaction
	consumerTx := app.Application.StartConsumer("consumer-trace")

	// Add a nested transaction
	txn := consumerTx.AddTxn("validate-order", logmanager.TxnTypeOther)
	txn.End()

	// End the consumer transaction
	consumerTx.End()

	// Assert logged entry count (both nested and parent transaction)
	assert.Equal(t, 2, app.CountLoggedEntries(), "Should have exactly two logged entries")

	// Get entries to check the parent transaction (last entry)
	entries := app.GetLoggedEntries()
	lastEntry := entries[1]

	// Verify parent consumer transaction fields
	assert.Equal(t, "consumer-trace", lastEntry.Data["trace_id"], "Should log correct trace_id")
	assert.Equal(t, "consumer", lastEntry.Data["name"], "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeConsumer, lastEntry.Data["type"], "Should log Consumer transaction type")
	assert.Equal(t, logrus.InfoLevel, lastEntry.Level, "Should log at Info level for successful consumer transactions")
	assert.Equal(t, "", lastEntry.Message, "Should have empty message for successful transactions")
}

func TestTxnRecord_End_CronTransaction(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	// Start a cron transaction using generic Start method
	cronTx := app.Application.Start("cron-trace", "cleanup-job", logmanager.TxnTypeCron)

	// Add some work
	txn := cronTx.AddDatabase("DELETE expired_sessions")
	txn.End()

	// End the cron transaction
	cronTx.End()

	// Assert logged entry count (both nested and parent transaction)
	assert.Equal(t, 2, app.CountLoggedEntries(), "Should have exactly two logged entries")

	// Get entries to check the parent transaction (last entry)
	entries := app.GetLoggedEntries()
	lastEntry := entries[1]

	// Verify parent cron transaction fields
	assert.Equal(t, "cron-trace", lastEntry.Data["trace_id"], "Should log correct trace_id")
	assert.Equal(t, "cleanup-job", lastEntry.Data["name"], "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeCron, lastEntry.Data["type"], "Should log Cron transaction type")
	assert.Equal(t, logrus.InfoLevel, lastEntry.Level, "Should log at Info level for successful cron transactions")
	assert.Equal(t, "", lastEntry.Message, "Should have empty message for successful transactions")
}

func TestTxnRecord_End_GrpcTransaction(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	parentTx := app.Application.StartHttp("grpc-trace", "parent")

	// Add a gRPC transaction
	grpcTxn := parentTx.AddTxn("UserService/GetUser", logmanager.TxnTypeGrpc)
	grpcTxn.SetResponseBodyAndCode([]byte(`{"user_id": "123", "name": "John"}`), 200)

	// End the gRPC transaction
	grpcTxn.End()

	// Assert logged entry count
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")

	// Verify gRPC transaction fields
	assert.True(t, app.HasLoggedField("trace_id"), "Should log trace_id field")
	assert.True(t, app.HasLoggedField("name"), "Should log name field")
	assert.True(t, app.HasLoggedField("type"), "Should log type field")
	assert.Equal(t, "grpc-trace", app.GetLoggedField("trace_id"), "Should log correct trace_id")
	assert.Equal(t, "UserService/GetUser", app.GetLoggedField("name"), "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeGrpc, app.GetLoggedField("type"), "Should log gRPC transaction type")

	// Check for response fields
	assert.True(t, app.HasLoggedField("response"), "Should log response body field")
	assert.True(t, app.HasLoggedField("status"), "Should log status code field")
	assert.Equal(t, 200, app.GetLoggedField("status"), "Should log correct status code")

	// Verify log level
	assert.Equal(t, logrus.InfoLevel, app.GetLoggedLevel(), "Should log at Info level for successful gRPC transactions")
	assert.Equal(t, "", app.GetLoggedMessage(), "Should have empty message for successful transactions")
}

func TestTxnRecord_End_OtherTransaction(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	parentTx := app.Application.StartHttp("other-trace", "parent")

	// Add an "other" type transaction (e.g., custom background task)
	otherTxn := parentTx.AddTxn("email-notification", logmanager.TxnTypeOther)

	// End the transaction
	otherTxn.End()

	// Assert logged entry count
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")

	// Verify other transaction fields
	assert.True(t, app.HasLoggedField("trace_id"), "Should log trace_id field")
	assert.True(t, app.HasLoggedField("name"), "Should log name field")
	assert.True(t, app.HasLoggedField("type"), "Should log type field")
	assert.Equal(t, "other-trace", app.GetLoggedField("trace_id"), "Should log correct trace_id")
	assert.Equal(t, "email-notification", app.GetLoggedField("name"), "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeOther, app.GetLoggedField("type"), "Should log Other transaction type")

	// Verify log level
	assert.Equal(t, logrus.InfoLevel, app.GetLoggedLevel(), "Should log at Info level for successful other transactions")
	assert.Equal(t, "", app.GetLoggedMessage(), "Should have empty message for successful transactions")
}

func TestTxnRecord_End_WithTags(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	parentTx := app.Application.StartHttp("tags-trace", "parent")
	txn := parentTx.AddTxn("process-payment", logmanager.TxnTypeApi)

	// Add tags
	txn.AddTags("payment", "api", "production")

	// End the transaction
	txn.End()

	// Assert logged entry count
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")

	// Verify standard transaction fields
	assert.Equal(t, "tags-trace", app.GetLoggedField("trace_id"), "Should log correct trace_id")
	assert.Equal(t, "process-payment", app.GetLoggedField("name"), "Should log correct transaction name")
	assert.Equal(t, logmanager.TxnTypeApi, app.GetLoggedField("type"), "Should log API transaction type")

	// Verify tags are logged
	assert.True(t, app.HasLoggedField("tags"), "Should log tags field")
	tags := app.GetLoggedField("tags").([]string)
	assert.Contains(t, tags, "payment", "Should contain payment tag")
	assert.Contains(t, tags, "api", "Should contain api tag")
	assert.Contains(t, tags, "production", "Should contain production tag")

	// Verify log level
	assert.Equal(t, logrus.InfoLevel, app.GetLoggedLevel(), "Should log at Info level for successful transactions")
	assert.Equal(t, "", app.GetLoggedMessage(), "Should have empty message for successful transactions")
}

func TestTxnRecord_End_MultipleSegments(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	// Start a main HTTP transaction
	mainTx := app.Application.StartHttp("multi-segment-trace", "order-processing")

	// Set a successful response to avoid default error status
	mainTx.SetResponseBodyAndCode([]byte(`{"status": "completed"}`), 200)

	// Add multiple different types of segments
	dbTxn := mainTx.AddDatabase("SELECT * FROM users WHERE id = ?")
	dbTxn.End()

	apiTxn := mainTx.AddTxn("POST /payment/process", logmanager.TxnTypeApi)
	apiTxn.SetResponseBodyAndCode([]byte(`{"status": "success"}`), 200)
	apiTxn.End()

	grpcTxn := mainTx.AddTxn("NotificationService/SendEmail", logmanager.TxnTypeGrpc)
	grpcTxn.End()

	otherTxn := mainTx.AddTxn("audit-log", logmanager.TxnTypeOther)
	otherTxn.End()

	// End the main transaction
	mainTx.End()

	// Assert logged entry count (4 segments + 1 main transaction = 5 total)
	assert.Equal(t, 5, app.CountLoggedEntries(), "Should have exactly five logged entries")

	// Get all entries
	entries := app.GetLoggedEntries()

	// Verify all transactions have the same trace_id
	for i, entry := range entries {
		assert.Equal(t, "multi-segment-trace", entry.Data["trace_id"],
			"Entry %d should have correct trace_id", i)
	}

	// Verify each segment type
	assert.Equal(t, "SELECT * FROM users WHERE id = ?", entries[0].Data["name"], "First entry should be database transaction")
	assert.Equal(t, logmanager.TxnTypeDatabase, entries[0].Data["type"], "First entry should be database type")

	assert.Equal(t, "POST /payment/process", entries[1].Data["name"], "Second entry should be API transaction")
	assert.Equal(t, logmanager.TxnTypeApi, entries[1].Data["type"], "Second entry should be API type")

	assert.Equal(t, "NotificationService/SendEmail", entries[2].Data["name"], "Third entry should be gRPC transaction")
	assert.Equal(t, logmanager.TxnTypeGrpc, entries[2].Data["type"], "Third entry should be gRPC type")

	assert.Equal(t, "audit-log", entries[3].Data["name"], "Fourth entry should be other transaction")
	assert.Equal(t, logmanager.TxnTypeOther, entries[3].Data["type"], "Fourth entry should be other type")

	// Verify main transaction (last entry)
	mainEntry := entries[4]
	assert.Equal(t, "order-processing", mainEntry.Data["name"], "Main transaction should have correct name")
	assert.Equal(t, logmanager.TxnTypeHttp, mainEntry.Data["type"], "Main transaction should be HTTP type")

	// Verify all entries logged at Info level (successful transactions)
	for i, entry := range entries {
		assert.Equal(t, logrus.InfoLevel, entry.Level, "Entry %d should be Info level", i)
		assert.Equal(t, "", entry.Message, "Entry %d should have empty message", i)
	}
}
