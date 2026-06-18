package lmgin_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"

	"github.com/gin-gonic/gin"
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
// internal work — each of which must produce its own log entry. The surrounding
// middleware produces the http entry on tx.End().
func fullProcessingHandler(c *gin.Context) {
	tx := logmanager.FromContext(c.Request.Context())

	// api: outbound HTTP call
	apiTxn := tx.AddTxn("call-payment-api", logmanager.TxnTypeApi)
	apiTxn.SetRequestValue(map[string]any{"order_id": c.Param("id")})
	apiTxn.SetResponseValue(map[string]any{"status": "charged"})
	apiTxn.End()

	// database: query
	dbTxn := logmanager.StartDatabaseSegment(tx, logmanager.DatabaseSegment{
		Name:  "select-orders",
		Table: "orders",
		Query: "SELECT * FROM orders WHERE id = ?",
		Host:  "localhost",
	})
	dbTxn.End()

	// other: internal computation
	otherTxn := logmanager.StartOtherSegment(tx, logmanager.OtherSegment{
		Name:  "compute-tax",
		Extra: map[string]interface{}{"region": "ID"},
	})
	otherTxn.End()

	c.JSON(http.StatusOK, gin.H{"message": "ok", "id": c.Param("id")})
}

func newRouter(app *logmanager.TestableApplication) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(lmgin.Middleware(app.Application))
	r.GET("/orders/:id", fullProcessingHandler)
	return r
}

// TestGinAllLogTypesPresent proves that a single REST request emits every
// expected transaction log type: http, api, database and other.
func TestGinAllLogTypesPresent(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()
	r := newRouter(app)

	req := httptest.NewRequest(http.MethodGet, "/orders/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

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

	// Every entry produced during one request must share the same trace_id.
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

// TestGinConcurrentRequests drives many requests concurrently against the same
// app/router to emulate high production traffic. Run with -race to surface any
// shared-state data races across requests.
func TestGinConcurrentRequests(t *testing.T) {
	const n = 200

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()
	r := newRouter(app)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/orders/123", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
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

// TestGinConcurrentFanoutWithinRequest is the key race probe: a single request
// whose handler fans out to many goroutines that all mutate the SAME transaction
// concurrently (parallel downstream calls — a common production pattern). This
// targets the shared txnRecords map, the tags slice and the attrs map. Run with
// -race; a "concurrent map writes" panic or a reported data race confirms the
// production risk.
func TestGinConcurrentFanoutWithinRequest(t *testing.T) {
	const fanout = 50

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(lmgin.Middleware(app.Application))
	r.GET("/fanout", func(c *gin.Context) {
		tx := logmanager.FromContext(c.Request.Context())

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

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fanout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.GreaterOrEqual(t, loggedTypes(app)[logmanager.TxnTypeHttp], 1, "http entry must be logged")
}
