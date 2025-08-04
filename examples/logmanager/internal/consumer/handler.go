package consumer

import (
	"context"
	"examples/logmanager/internal/publisher"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmrabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type Handler struct {
	App       *logmanager.Application
	Publisher *publisher.RabbitMQ
}

func (h *Handler) StartConsumer(d amqp091.Delivery) error {
	// CorrelationId as your trace ID
	tx := h.App.StartConsumer(d.CorrelationId)
	tx.SetConsumer(lmrabbitmq.NewConsumer("hello", d))
	defer tx.End()

	// store log manager to context
	ctx := tx.ToContext(context.Background())
	// do your logic here
	insertData(ctx, "ok")

	err := h.Publisher.Publish(ctx,
		map[string]any{
			"foo": "bar",
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func insertData(ctx context.Context, value string) {
	txn := logmanager.StartDatabaseSegment(
		logmanager.FromContext(ctx),
		logmanager.DatabaseSegment{
			Table: "table_name",
			Query: "insert into table_name values " + value,
			Host:  "localhost",
		},
	)
	defer txn.End()

	// your logic here ...
	// it takes 200ms latency
	time.Sleep(200 * time.Millisecond)
}
