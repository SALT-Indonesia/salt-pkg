package lmrabbitmq_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmrabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConsumer(t *testing.T) {
	tests := []struct {
		name          string
		queue         string
		delivery      amqp091.Delivery
		expectedQueue string
		expectedExch  string
		expectedKey   string
		expectedBody  []byte
	}{
		{
			name:          "valid consumer creation",
			queue:         "test-queue",
			delivery:      amqp091.Delivery{Exchange: "test-exchange", RoutingKey: "test-key", Body: []byte("test-body")},
			expectedQueue: "test-queue",
			expectedExch:  "test-exchange",
			expectedKey:   "test-key",
			expectedBody:  []byte("test-body"),
		},
		{
			name:          "empty queue and delivery fields",
			queue:         "",
			delivery:      amqp091.Delivery{Exchange: "", RoutingKey: "", Body: []byte("")},
			expectedQueue: "",
			expectedExch:  "default",
			expectedKey:   "",
			expectedBody:  []byte(""),
		},
		{
			name:          "only queue provided",
			queue:         "only-queue",
			delivery:      amqp091.Delivery{Exchange: "", RoutingKey: "", Body: nil},
			expectedQueue: "only-queue",
			expectedExch:  "default",
			expectedKey:   "",
			expectedBody:  nil,
		},
		{
			name:          "delivery with large body",
			queue:         "large-body-queue",
			delivery:      amqp091.Delivery{Exchange: "large-exchange", RoutingKey: "large-key", Body: []byte("a very large body of message")},
			expectedQueue: "large-body-queue",
			expectedExch:  "large-exchange",
			expectedKey:   "large-key",
			expectedBody:  []byte("a very large body of message"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := lmrabbitmq.NewConsumer(tt.queue, tt.delivery)
			consumer.Init()

			assert.NotNil(t, consumer)
			assert.Equal(t, tt.expectedQueue, consumer.Queue)
			assert.Equal(t, tt.expectedExch, consumer.Exchange)
			assert.Equal(t, tt.expectedKey, consumer.RoutingKey)
			assert.Equal(t, tt.expectedBody, consumer.RequestBody)
		})
	}
}
