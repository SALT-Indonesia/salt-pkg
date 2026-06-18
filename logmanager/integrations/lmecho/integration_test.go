package lmecho_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmecho"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// loggedTypes counts the captured log entries grouped by their "type" field.
// Entries are produced both by the middleware (http) and by every nested
// TxnRecord.End() call (api, database, other, ...).
func loggedTypes(app *logmanager.TestableApplication) map[logmanager.TxnType]int {
	counts := make(map[logmanager.TxnType]int)
	for _, entry := range app.GetLoggedEntries() {
		if v, ok := entry.Data["type"]; ok {
			if t, ok := v.(logmanager.TxnType); ok {
				counts[t]++
			}
		}
	}
	return counts
}

// fullProcessingHandler simulates a realistic REST handler that, while serving
// one request, performs an outbound API call, a database query and some other
// internal work — each producing its own log entry. The middleware produces the
// http entry on tx.End().
func fullProcessingHandler(c echo.Context) error {
	tx := logmanager.FromContext(c.Request().Context())

	apiTxn := tx.AddTxn("call-payment-api", logmanager.TxnTypeApi)
	apiTxn.SetRequestValue(map[string]any{"order_id": c.Param("id")})
	apiTxn.SetResponseValue(map[string]any{"status": "charged"})
	apiTxn.End()

	dbTxn := logmanager.StartDatabaseSegment(tx, logmanager.DatabaseSegment{
		Name:  "select-orders",
		Table: "orders",
		Query: "SELECT * FROM orders WHERE id = ?",
		Host:  "localhost",
	})
	dbTxn.End()

	otherTxn := logmanager.StartOtherSegment(tx, logmanager.OtherSegment{
		Name:  "compute-tax",
		Extra: map[string]interface{}{"region": "ID"},
	})
	otherTxn.End()

	return c.JSON(http.StatusOK, map[string]any{"message": "ok", "id": c.Param("id")})
}

func newRouter(app *logmanager.TestableApplication) *echo.Echo {
	e := echo.New()
	e.Use(lmecho.Middleware(app.Application))
	e.GET("/orders/:id", fullProcessingHandler)
	return e
}

// TestEchoAllLogTypesPresent proves that a single REST request emits every
// expected transaction log type: http, api, database and other.
func TestEchoAllLogTypesPresent(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()
	e := newRouter(app)

	req := httptest.NewRequest(http.MethodGet, "/orders/123", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	counts := loggedTypes(app)
	for _, want := range []logmanager.TxnType{
		logmanager.TxnTypeHttp,
		logmanager.TxnTypeApi,
		logmanager.TxnTypeDatabase,
		logmanager.TxnTypeOther,
	} {
		assert.GreaterOrEqual(t, counts[want], 1, "expected at least one %q log entry", want)
	}

	var traceID string
	for _, entry := range app.GetLoggedEntries() {
		got, _ := entry.Data["trace_id"].(string)
		assert.NotEmpty(t, got, "trace_id should not be empty")
		if traceID == "" {
			traceID = got
		} else {
			assert.Equal(t, traceID, got, "all entries in a request should share trace_id")
		}
	}
}

// TestEchoConcurrentRequests drives many requests concurrently against the same
// app/router to emulate high production traffic. Run with -race to surface any
// shared-state data races across requests.
func TestEchoConcurrentRequests(t *testing.T) {
	const n = 200

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()
	e := newRouter(app)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/orders/123", nil)
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}()
	}
	wg.Wait()

	counts := loggedTypes(app)
	assert.Equal(t, n, counts[logmanager.TxnTypeHttp], "one http entry per request")
	assert.Equal(t, n, counts[logmanager.TxnTypeApi], "one api entry per request")
	assert.Equal(t, n, counts[logmanager.TxnTypeDatabase], "one database entry per request")
	assert.Equal(t, n, counts[logmanager.TxnTypeOther], "one other entry per request")
}

// TestEchoConcurrentFanoutWithinRequest is the key race probe: a single request
// whose handler fans out to many goroutines that all mutate the SAME transaction
// concurrently (parallel downstream calls). Targets the shared txnRecords map,
// tags slice and attrs map. Run with -race.
func TestEchoConcurrentFanoutWithinRequest(t *testing.T) {
	const fanout = 50

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	e := echo.New()
	e.Use(lmecho.Middleware(app.Application))
	e.GET("/fanout", func(c echo.Context) error {
		tx := logmanager.FromContext(c.Request().Context())

		var wg sync.WaitGroup
		wg.Add(fanout)
		for i := 0; i < fanout; i++ {
			go func(i int) {
				defer wg.Done()

				db := tx.AddDatabase("db-call")
				db.AddTags("db", "fanout")
				db.SetResponseValue(map[string]any{"row": i})
				db.End()

				api := tx.AddTxn("api-call", logmanager.TxnTypeApi)
				api.AddTags("api")
				api.SetResponseValue(map[string]any{"ok": true})
				api.End()
			}(i)
		}
		wg.Wait()

		return c.JSON(http.StatusOK, map[string]any{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fanout", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.GreaterOrEqual(t, loggedTypes(app)[logmanager.TxnTypeHttp], 1, "http entry must be logged")
}
