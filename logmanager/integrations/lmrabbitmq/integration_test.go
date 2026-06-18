package lmrabbitmq_test

import (
	"sync"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmrabbitmq"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// loggedTypes counts the captured log entries grouped by their "type" field.
// Entries are produced both by the consumer transaction (consumer) and by
// every nested TxnRecord.End() call (api, database, other, ...).
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

func newDelivery() amqp091.Delivery {
	return amqp091.Delivery{
		Exchange:   "orders-exchange",
		RoutingKey: "orders.created",
		Body:       []byte(`{"order_id":"123"}`),
	}
}

// processMessage simulates a realistic RabbitMQ consumer that, while handling
// one message, performs an outbound API call, a database query and some other
// internal work — each producing its own log entry. The consumer transaction
// produces the consumer entry on tx.End().
func processMessage(app *logmanager.Application, traceID string) {
	consumer := lmrabbitmq.NewConsumer("orders-queue", newDelivery())

	tx := app.StartConsumer(traceID)
	tx.SetConsumer(consumer)
	defer tx.End()

	apiTxn := tx.AddTxn("call-payment-api", logmanager.TxnTypeApi)
	apiTxn.SetRequestValue(map[string]any{"order_id": "123"})
	apiTxn.SetResponseValue(map[string]any{"status": "charged"})
	apiTxn.End()

	dbTxn := logmanager.StartDatabaseSegment(tx, logmanager.DatabaseSegment{
		Name:  "insert-order",
		Table: "orders",
		Query: "INSERT INTO orders ...",
		Host:  "localhost",
	})
	dbTxn.End()

	otherTxn := logmanager.StartOtherSegment(tx, logmanager.OtherSegment{
		Name:  "compute-tax",
		Extra: map[string]interface{}{"region": "ID"},
	})
	otherTxn.End()
}

// TestRabbitMQAllLogTypesPresent proves that handling a single message emits
// every expected transaction log type: consumer, api, database and other.
func TestRabbitMQAllLogTypesPresent(t *testing.T) {
	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	processMessage(app.Application, "trace-abc")

	counts := loggedTypes(app)
	for _, want := range []logmanager.TxnType{
		logmanager.TxnTypeConsumer,
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

// TestRabbitMQConcurrentMessages handles many messages concurrently against the
// same app to emulate high-throughput consumption. Run with -race to surface any
// shared-state data races across messages.
func TestRabbitMQConcurrentMessages(t *testing.T) {
	const n = 200

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			processMessage(app.Application, "trace-abc")
		}()
	}
	wg.Wait()

	counts := loggedTypes(app)
	assert.Equal(t, n, counts[logmanager.TxnTypeConsumer], "one consumer entry per message")
	assert.Equal(t, n, counts[logmanager.TxnTypeApi], "one api entry per message")
	assert.Equal(t, n, counts[logmanager.TxnTypeDatabase], "one database entry per message")
	assert.Equal(t, n, counts[logmanager.TxnTypeOther], "one other entry per message")
}

// TestRabbitMQConcurrentFanoutWithinMessage is the key race probe: a single
// message whose handler fans out to many goroutines that all mutate the SAME
// transaction concurrently. Targets the shared txnRecords map, tags slice and
// attrs map. Run with -race.
func TestRabbitMQConcurrentFanoutWithinMessage(t *testing.T) {
	const fanout = 50

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	consumer := lmrabbitmq.NewConsumer("orders-queue", newDelivery())
	tx := app.StartConsumer("trace-abc")
	tx.SetConsumer(consumer)

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
	tx.End()

	assert.GreaterOrEqual(t, loggedTypes(app)[logmanager.TxnTypeConsumer], 1, "consumer entry must be logged")
}
