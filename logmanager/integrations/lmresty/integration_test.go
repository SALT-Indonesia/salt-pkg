package lmresty_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// loggedTypes counts the captured log entries grouped by their "type" field.
// Entries are produced by the root transaction (http) and by every nested
// TxnRecord.End() call, including the resty api segment.
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

func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"success"}`))
	}))
}

// handleRequest simulates a realistic flow: a root transaction makes an outbound
// HTTP call through resty (logged via lmresty.NewTxn as an api segment), plus a
// database query and some other internal work. Every step produces a log entry;
// the root transaction produces the http entry on tx.End().
func handleRequest(app *logmanager.Application, client *resty.Client, traceID string) error {
	tx := app.StartHttp(traceID, "GET /proxy")
	defer tx.End()

	ctx := tx.ToContext(context.Background())

	resp, err := client.R().SetContext(ctx).Get("/downstream")
	if err != nil {
		return err
	}
	// api: the outbound resty call
	apiTxn := lmresty.NewTxn(resp)
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

	return nil
}

// TestRestyAllLogTypesPresent proves that a single request flow emits every
// expected transaction log type: http, api (resty), database and other.
func TestRestyAllLogTypesPresent(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	assert.NoError(t, handleRequest(app.Application, client, "trace-abc"))

	counts := loggedTypes(app)
	for _, want := range []logmanager.TxnType{
		logmanager.TxnTypeHttp,
		logmanager.TxnTypeApi,
		logmanager.TxnTypeDatabase,
		logmanager.TxnTypeOther,
	} {
		assert.GreaterOrEqual(t, counts[want], 1, "expected at least one %q log entry", want)
	}

	for _, entry := range app.GetLoggedEntries() {
		assert.Equal(t, "trace-abc", entry.Data["trace_id"], "all entries should share trace_id")
	}
}

// TestRestyConcurrentRequests drives many request flows concurrently against the
// same app and resty client to emulate high production traffic. Run with -race
// to surface any shared-state data races across requests.
func TestRestyConcurrentRequests(t *testing.T) {
	const n = 200

	server := newTestServer()
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			assert.NoError(t, handleRequest(app.Application, client, "trace-abc"))
		}()
	}
	wg.Wait()

	counts := loggedTypes(app)
	assert.Equal(t, n, counts[logmanager.TxnTypeHttp], "one http entry per request")
	assert.Equal(t, n, counts[logmanager.TxnTypeApi], "one api entry per request")
	assert.Equal(t, n, counts[logmanager.TxnTypeDatabase], "one database entry per request")
	assert.Equal(t, n, counts[logmanager.TxnTypeOther], "one other entry per request")
}

// TestRestyConcurrentFanoutWithinRequest is the key race probe: a single request
// whose handler fans out to many goroutines that each make a resty call and add
// segments to the SAME transaction concurrently. Targets the shared txnRecords
// map, tags slice and attrs map. Run with -race.
func TestRestyConcurrentFanoutWithinRequest(t *testing.T) {
	const fanout = 50

	server := newTestServer()
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	tx := app.StartHttp("trace-abc", "GET /proxy")
	ctx := tx.ToContext(context.Background())

	var wg sync.WaitGroup
	wg.Add(fanout)
	for i := 0; i < fanout; i++ {
		go func(i int) {
			defer wg.Done()

			resp, err := client.R().SetContext(ctx).Get("/downstream")
			assert.NoError(t, err)
			apiTxn := lmresty.NewTxn(resp)
			apiTxn.AddTags("api", "fanout")
			apiTxn.End()

			db := tx.AddDatabase("db-call")
			db.SetResponseValue(map[string]any{"row": i})
			db.End()
		}(i)
	}
	wg.Wait()
	tx.End()

	assert.GreaterOrEqual(t, loggedTypes(app)[logmanager.TxnTypeHttp], 1, "http entry must be logged")
}
