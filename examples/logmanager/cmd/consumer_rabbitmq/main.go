package main

import (
	"examples/logmanager/internal/consumer"
	"examples/logmanager/internal/publisher"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/rabbitmq/amqp091-go"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	deliveries, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	// start to initialize application here
	app := logmanager.NewApplication()
	h := &consumer.Handler{
		App:       app,
		Publisher: &publisher.RabbitMQ{Channel: ch, Key: "hello_key"},
	}

	go func() {
		for d := range deliveries {
			err = h.StartConsumer(d)
			if err != nil {
				_ = d.Reject(true)
			}

			_ = d.Ack(false)
		}
	}()

	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
