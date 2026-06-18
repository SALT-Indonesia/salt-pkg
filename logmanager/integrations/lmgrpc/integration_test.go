package lmgrpc_test

import (
	"context"
	"sync"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// loggedTypes counts the captured log entries grouped by their "type" field.
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

// fullProcessingHandler simulates a gRPC handler that, while serving one RPC,
// performs an outbound API call, a database query, some other internal work and
// a downstream gRPC call — each producing its own log entry. The interceptor
// produces the http entry on tx.End().
func fullProcessingHandler(ctx context.Context, _ interface{}) (interface{}, error) {
	tx := logmanager.FromContext(ctx)

	// api: outbound HTTP call
	apiTxn := tx.AddTxn("call-payment-api", logmanager.TxnTypeApi)
	apiTxn.SetRequestValue(map[string]any{"order_id": "123"})
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

	// grpc: downstream gRPC call
	grpcTxn := logmanager.StartGrpcSegment(tx, logmanager.GrpcSegment{
		Url:     "/payment.Service/Charge",
		Request: map[string]any{"order_id": "123"},
	})
	grpcTxn.End()

	return "mock-response", nil
}

func newServerCall(app *logmanager.TestableApplication, handler grpc.UnaryHandler) (interface{}, error) {
	interceptor := lmgrpc.UnaryServerInterceptor(app.Application)
	md := metadata.Pairs(app.TraceIDHeaderKey(), "trace-abc")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	return interceptor(ctx, "mock-request", &grpc.UnaryServerInfo{FullMethod: "/order.Service/Get"}, handler)
}

// TestGrpcAllLogTypesPresent proves that a single gRPC unary call emits every
// expected transaction log type: http, api, database, other and grpc.
func TestGrpcAllLogTypesPresent(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	resp, err := newServerCall(app, fullProcessingHandler)
	assert.NoError(t, err)
	assert.Equal(t, "mock-response", resp)

	counts := loggedTypes(app)
	for _, want := range []logmanager.TxnType{
		logmanager.TxnTypeHttp,
		logmanager.TxnTypeApi,
		logmanager.TxnTypeDatabase,
		logmanager.TxnTypeOther,
		logmanager.TxnTypeGrpc,
	} {
		assert.GreaterOrEqual(t, counts[want], 1, "expected at least one %q log entry", want)
	}

	// Every entry produced during one RPC must share the same trace_id.
	for _, entry := range app.GetLoggedEntries() {
		assert.Equal(t, "trace-abc", entry.Data["trace_id"], "all entries should share trace_id")
	}
}

// TestGrpcConcurrentCalls drives many unary calls concurrently against the same
// app to emulate high production traffic. Run with -race to surface any
// shared-state data races across calls.
func TestGrpcConcurrentCalls(t *testing.T) {
	const n = 200

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			resp, err := newServerCall(app, fullProcessingHandler)
			assert.NoError(t, err)
			assert.Equal(t, "mock-response", resp)
		}()
	}
	wg.Wait()

	counts := loggedTypes(app)
	assert.Equal(t, n, counts[logmanager.TxnTypeHttp], "one http entry per call")
	assert.Equal(t, n, counts[logmanager.TxnTypeApi], "one api entry per call")
	assert.Equal(t, n, counts[logmanager.TxnTypeDatabase], "one database entry per call")
	assert.Equal(t, n, counts[logmanager.TxnTypeOther], "one other entry per call")
	assert.Equal(t, n, counts[logmanager.TxnTypeGrpc], "one grpc entry per call")
}

// TestGrpcConcurrentFanoutWithinCall is the key race probe: a single RPC whose
// handler fans out to many goroutines that all mutate the SAME transaction
// concurrently. This targets the shared txnRecords map, the tags slice and the
// attrs map. Run with -race; a "concurrent map writes" panic or a reported data
// race confirms the production risk.
func TestGrpcConcurrentFanoutWithinCall(t *testing.T) {
	const fanout = 50

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	handler := func(ctx context.Context, _ interface{}) (interface{}, error) {
		tx := logmanager.FromContext(ctx)

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

		return "mock-response", nil
	}

	resp, err := newServerCall(app, handler)
	assert.NoError(t, err)
	assert.Equal(t, "mock-response", resp)
	assert.GreaterOrEqual(t, loggedTypes(app)[logmanager.TxnTypeHttp], 1, "http entry must be logged")
}
