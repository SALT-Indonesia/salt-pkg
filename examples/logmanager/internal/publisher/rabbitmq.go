package publisher

import (
	"context"
	"encoding/json"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Channel *amqp091.Channel
	Key     string
}

func (r *RabbitMQ) Publish(ctx context.Context, payload map[string]any) error {
	data, _ := json.Marshal(payload)

	tx := logmanager.FromContext(ctx)
	txn := logmanager.StartOtherSegment(
		tx,
		logmanager.OtherSegment{
			Name: "rabbitmq/hello_key",
			Extra: map[string]any{
				"data":     string(data),
				"key":      r.Key,
				"exchange": "amq.direct",
			},
		},
	)
	defer txn.End()

	err := r.Channel.PublishWithContext(
		ctx,
		"amq.direct",
		r.Key,
		false,
		false,
		amqp091.Publishing{
			CorrelationId: tx.TraceID(),
			ContentType:   "application/json",
			Body:          data,
		},
	)
	if err != nil {
		txn.NoticeError(err) // log error
	}

	return err
}
