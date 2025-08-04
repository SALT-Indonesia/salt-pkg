package eventmanager

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmrabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	checker = make(map[string]bool)
	counter = make(map[string]uint)
)

func refresh() {
	for {
		time.Sleep(refreshDelay)
		checker = make(map[string]bool)
		counter = make(map[string]uint)
	}
}

func requeue(msg amqp.Delivery) {
	if counter[msg.CorrelationId] < maxRetry {
		time.Sleep(retryDelay)
		_ = msg.Nack(false, true)
		counter[msg.CorrelationId]++
	}
}

func logHandleMessage(ctx context.Context, msg amqp.Delivery, queueName string) (*logmanager.Transaction, context.Context) {
	tx := app.StartConsumer(msg.CorrelationId)
	tx.SetConsumer(lmrabbitmq.NewConsumer(queueName, msg))

	return tx, tx.ToContext(ctx)
}

func handleMessage[Message any](ctx context.Context, msg amqp.Delivery, queueName string, handlers []Handler[Message]) {
	if !checker[msg.CorrelationId] {
		checker[msg.CorrelationId] = true

		tx, ctx := logHandleMessage(ctx, msg, queueName)
		defer tx.End()

		for _, handler := range handlers {
			var body Message
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				logmanager.LogErrorWithContext(ctx, err)
			}

			domainErr, infrastructureErr := handler(body)
			if domainErr != nil {
				logmanager.LogErrorWithContext(ctx, domainErr)
			}
			if infrastructureErr != nil {
				logmanager.LogErrorWithContext(ctx, infrastructureErr)

				requeue(msg)
			}
		}
	}

	_ = msg.Ack(false)
}

func handleMessages[Message any](ctx context.Context, msgs <-chan amqp.Delivery, queueName string, handlers []Handler[Message]) {
	for msg := range msgs {
		go handleMessage(ctx, msg, queueName, handlers)
	}
}

func reconnect[Message any](ctx context.Context, queueName, exchange string, handlers []Handler[Message]) {
	time.Sleep(delay)
	subscribe(ctx, queueName, exchange, handlers)
}

func resubscribe[Message any](notify chan *amqp.Error, ctx context.Context, queueName, exchange string, handlers []Handler[Message]) {
	if err := <-notify; err != nil {
		logmanager.LogErrorWithContext(ctx, err)

		reconnect(ctx, queueName, exchange, handlers)
	}
}

func channelling[Message any](ctx context.Context, connection *amqp.Connection, queueName, exchange string, handlers []Handler[Message]) {
	channel, err := connection.Channel()
	if err != nil {
		logmanager.LogErrorWithContext(ctx, err)

		return
	}

	if err := channel.Qos(prefetchCount, 0, false); err != nil { // 10 is prefetch count
		logmanager.LogErrorWithContext(ctx, err)

		return
	}

	if err = channel.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		logmanager.LogErrorWithContext(ctx, err)

		return
	}

	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logmanager.LogErrorWithContext(ctx, err)

		return
	}

	if err := channel.QueueBind(queue.Name, exchange, exchange, false, nil); err != nil {
		logmanager.LogErrorWithContext(ctx, err)

		return
	}

	msgs, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		logmanager.LogErrorWithContext(ctx, err)

		return
	}

	go handleMessages(ctx, msgs, queueName, handlers)
}

func subscribe[Message any](ctx context.Context, queueName, exchange string, handlers []Handler[Message]) {
	connection, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		logmanager.LogErrorWithContext(ctx, err)

		reconnect(ctx, exchange, queueName, handlers)

		return
	}

	notify := connection.NotifyClose(make(chan *amqp.Error))
	go resubscribe(notify, ctx, queueName, exchange, handlers)

	for range channels {
		go channelling(ctx, connection, queueName, exchange, handlers)
	}

	<-ctx.Done()
}

func Subscribe[Message any](ctx context.Context, queueName, exchange string, handlers []Handler[Message]) {
	go refresh()
	if app == nil {
		app = logmanager.NewApplication(logmanager.WithAppName(queueName))
	}
	setVars()
	subscribe(ctx, queueName, exchange, handlers)
}
