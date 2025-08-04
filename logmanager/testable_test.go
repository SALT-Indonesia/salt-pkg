package logmanager_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewTestableApplication(t *testing.T) {
	app := logmanager.NewTestableApplication()

	assert.NotNil(t, app)
	assert.NotNil(t, app.Application)

	// Test that we can start transactions
	tx := app.StartHttp("test-trace-id", "test-transaction")
	assert.NotNil(t, tx)
	assert.Equal(t, "test-trace-id", tx.TraceID())
}

func TestTestableApplication_LoggedEntries(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Initially no entries
	assert.Equal(t, 0, app.CountLoggedEntries())
	assert.Nil(t, app.GetLastLoggedEntry())

	// Start a transaction and end it to trigger logging
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.End()

	// Should have one logged entry
	assert.Equal(t, 1, app.CountLoggedEntries())
	assert.NotNil(t, app.GetLastLoggedEntry())
}

func TestTestableApplication_LoggedFields(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Start a transaction and end it
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.End()

	// Check logged fields
	fields := app.GetLoggedFields()
	assert.NotNil(t, fields)

	// Verify expected fields are present
	assert.True(t, app.HasLoggedField("trace_id"))
	assert.True(t, app.HasLoggedField("name"))
	assert.True(t, app.HasLoggedField("type"))
	assert.True(t, app.HasLoggedField("start"))
	assert.True(t, app.HasLoggedField("latency"))
	assert.True(t, app.HasLoggedField("service"))

	// Verify field values
	assert.Equal(t, "test-trace-id", app.GetLoggedField("trace_id"))
	assert.Equal(t, "test-transaction", app.GetLoggedField("name"))
	assert.Equal(t, logmanager.TxnTypeHttp, app.GetLoggedField("type"))
	assert.Equal(t, "default", app.GetLoggedField("service"))
}

func TestTestableApplication_LoggedLevel(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Test successful transaction with proper response code (should log at Info level)
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.SetResponseBodyAndCode([]byte("success"), 200)
	tx.End()

	assert.Equal(t, logrus.InfoLevel, app.GetLoggedLevel())
	assert.Equal(t, "", app.GetLoggedMessage()) // Info logs have empty message
}

func TestTestableApplication_LoggedError(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Test transaction with error (set a success response code to avoid internal server error)
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.SetResponseBodyAndCode([]byte("error"), 200) // Set response first
	tx.NoticeError(errors.New("test error"))
	tx.End()

	assert.Equal(t, logrus.ErrorLevel, app.GetLoggedLevel())
	assert.Equal(t, "test error", app.GetLoggedMessage())
}

func TestTestableApplication_LoggedBusinessError(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Test transaction with business error (set a success response code to avoid internal server error)
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.SetResponseBodyAndCode([]byte("business error"), 200) // Set response first
	tx.SetBusinessError(errors.New("business error"))
	tx.End()

	assert.Equal(t, logrus.WarnLevel, app.GetLoggedLevel())
	assert.Equal(t, "business error", app.GetLoggedMessage())
}

func TestTestableApplication_HttpRequestLogging(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Create a test HTTP request
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/users?id=123", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")

	// Start transaction and set request
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.SetWebRequest(req)
	tx.SetResponseBodyAndCode([]byte("success"), 200) // Add response to avoid error logging
	tx.End()

	// Verify request-related fields are logged (using actual field names from JSON output)
	assert.True(t, app.HasLoggedField("method"))
	assert.True(t, app.HasLoggedField("url"))
	assert.True(t, app.HasLoggedField("query_param"))
	assert.True(t, app.HasLoggedField("host"))

	assert.Equal(t, "POST", app.GetLoggedField("method"))
	assert.Equal(t, "http://example.com/api/users", app.GetLoggedField("url"))
	assert.Equal(t, "example.com", app.GetLoggedField("host"))

	// Check query parameters
	queryParam := app.GetLoggedField("query_param")
	assert.NotNil(t, queryParam)
}

func TestTestableApplication_HttpResponseLogging(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Create a test HTTP response
	resp := &http.Response{
		StatusCode: 201,
	}

	// Start transaction and set response
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.SetResponse(resp)
	tx.End()

	// Verify response-related fields are logged (using actual field name from JSON output)
	assert.True(t, app.HasLoggedField("status"))
	assert.Equal(t, 201, app.GetLoggedField("status"))
}

func TestTestableApplication_MultipleEntries(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Create multiple transactions
	// First transaction - successful (should log as Info)
	tx1 := app.StartHttp("trace-1", "transaction-1")
	tx1.SetResponseBodyAndCode([]byte("success"), 200)
	tx1.End()

	// Second transaction - with error (should log as Error)
	tx2 := app.StartHttp("trace-2", "transaction-2")
	tx2.SetResponseBodyAndCode([]byte("error"), 200) // Set response first
	tx2.NoticeError(errors.New("error in tx2"))
	tx2.End()

	// Should have 2 entries
	assert.Equal(t, 2, app.CountLoggedEntries())

	// Get entries by level
	infoEntries := app.GetLoggedEntriesWithLevel(logrus.InfoLevel)
	errorEntries := app.GetLoggedEntriesWithLevel(logrus.ErrorLevel)

	assert.Equal(t, 1, len(infoEntries))
	assert.Equal(t, 1, len(errorEntries))

	// Verify the info entry
	assert.Equal(t, "trace-1", infoEntries[0].Data["trace_id"])
	assert.Equal(t, "transaction-1", infoEntries[0].Data["name"])

	// Verify the error entry
	assert.Equal(t, "trace-2", errorEntries[0].Data["trace_id"])
	assert.Equal(t, "transaction-2", errorEntries[0].Data["name"])
	assert.Equal(t, "error in tx2", errorEntries[0].Message)
}

func TestTestableApplication_FilterByField(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Create transactions with different trace IDs
	tx1 := app.StartHttp("trace-1", "transaction-1")
	tx1.End()

	tx2 := app.StartHttp("trace-2", "transaction-2")
	tx2.End()

	// Get entries with trace_id field
	entriesWithTraceId := app.GetLoggedEntriesWithField("trace_id")
	assert.Equal(t, 2, len(entriesWithTraceId))

	// Get entries with a field that doesn't exist
	entriesWithNonExistentField := app.GetLoggedEntriesWithField("non_existent_field")
	assert.Equal(t, 0, len(entriesWithNonExistentField))
}

func TestTestableApplication_ResetEntries(t *testing.T) {
	app := logmanager.NewTestableApplication()

	// Create a transaction
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.End()

	// Should have one entry
	assert.Equal(t, 1, app.CountLoggedEntries())

	// Reset entries
	app.ResetLoggedEntries()

	// Should have no entries
	assert.Equal(t, 0, app.CountLoggedEntries())
	assert.Nil(t, app.GetLastLoggedEntry())
}

func TestApplication_AddTestHook(t *testing.T) {
	// Test adding test hook to existing application
	app := logmanager.NewApplication()
	testHook := app.AddTestHook()

	assert.NotNil(t, testHook)

	// Create a transaction and verify it's captured
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.End()

	// Should have captured the log entry
	assert.Equal(t, 1, len(testHook.AllEntries()))
	lastEntry := testHook.LastEntry()
	assert.NotNil(t, lastEntry)
	assert.Equal(t, "test-trace-id", lastEntry.Data["trace_id"])
}

func TestApplication_AddTestHook_NilApplication(t *testing.T) {
	var app *logmanager.Application
	testHook := app.AddTestHook()

	assert.Nil(t, testHook)
}

func TestTestableApplication_WithOptions(t *testing.T) {
	app := logmanager.NewTestableApplication(
		logmanager.WithAppName("test-service"),
		logmanager.WithTags("tag1", "tag2"),
	)

	// Create a transaction
	tx := app.StartHttp("test-trace-id", "test-transaction")
	tx.End()

	// Verify options are reflected in logged fields
	assert.Equal(t, "test-service", app.GetLoggedField("service"))
	assert.Equal(t, []string{"tag1", "tag2"}, app.GetLoggedField("tags"))
}
