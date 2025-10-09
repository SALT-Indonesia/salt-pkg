package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/rabbitmq/amqp091-go"
)

type MessageHandler struct {
	app *logmanager.Application
	ch  *amqp091.Channel
}

type Message struct {
	ID      string    `json:"id"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

func main() {
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	app := logmanager.NewApplication(
		logmanager.WithAppName("rabbitmq-consumer"),
	)

	handler := &MessageHandler{
		app: app,
		ch:  ch,
	}

	if err := handler.setupQueue(); err != nil {
		log.Fatalf("Failed to setup queue: %v", err)
	}

	if err := handler.startConsumer(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
}

func (h *MessageHandler) setupQueue() error {
	_, err := h.ch.QueueDeclare(
		"hello", // queue name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	return err
}

func (h *MessageHandler) startConsumer() error {
	deliveries, err := h.ch.Consume(
		"hello", // queue
		"",      // consumer
		true,    // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for delivery := range deliveries {
			h.processMessage(delivery)
		}
	}()

	fmt.Println("RabbitMQ consumer started. Press CTRL+C to exit")
	<-forever

	return nil
}

func (h *MessageHandler) processMessage(delivery amqp091.Delivery) {
	txn := h.app.Start(
		fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		"rabbitmq-consumer",
		logmanager.TxnTypeConsumer,
	)
	defer txn.End()

	var msg Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		txn.NoticeError(err)
		return
	}

	fmt.Printf("Processed message: %s\n", msg.Content)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
}