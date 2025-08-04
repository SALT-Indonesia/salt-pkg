package lmrabbitmq

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/rabbitmq/amqp091-go"
)

// NewConsumer creates a new Consumer instance with the specified queue name and delivery details from amqp091.Delivery.
func NewConsumer(queue string, d amqp091.Delivery) *logmanager.Consumer {
	c := &logmanager.Consumer{
		Queue:       queue,
		Exchange:    d.Exchange,
		RoutingKey:  d.RoutingKey,
		RequestBody: d.Body,
	}
	c.Init()
	return c
}
