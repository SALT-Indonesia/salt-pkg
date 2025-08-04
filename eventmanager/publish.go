package eventmanager

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func validate(body any) error {
	if err := val.Struct(body); err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if !errors.As(err, &invalidValidationError) {
			return err
		}
	}
	return nil
}

func publish(ctx context.Context, exchange string, body any) error {
	connection, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		return err
	}
	defer func() {
		_ = connection.Close()
	}()

	channel, _ := connection.Channel()
	defer func() {
		_ = channel.Close()
	}()

	if err = channel.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	tx := logmanager.FromContext(ctx)
	if tx == nil {
		return errors.New("transaction from the request context cannot be empty")
	}
	txn := logmanager.StartOtherSegment(
		tx,
		logmanager.OtherSegment{
			Name: "rabbitmq/hello_key",
			Extra: map[string]any{
				"data":     string(payload),
				"key":      exchange,
				"exchange": exchange,
			},
		},
	)
	defer txn.End()

	correlationID := tx.TraceID()
	if correlationID == "" {
		correlationID = uuid.New().String()
	}
	_ = channel.PublishWithContext(
		ctx,
		exchange,
		exchange,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			DeliveryMode:  amqp.Persistent,
			CorrelationId: correlationID,
			Body:          payload,
		},
	)

	return nil
}

func Publish(ctx context.Context, exchange string, body any) error {
	setVars()
	if err := validate(body); err != nil {
		return err
	}
	return publish(ctx, exchange, body)
}
