package logmanager

import "github.com/SALT-Indonesia/salt-pkg/logmanager/internal"

type Consumer struct {

	// Queue specifies the name of the queue to which the consumer is subscribed.
	Queue string

	// RoutingKey specifies the routing key used for message filtering in messaging systems like RabbitMQ.
	RoutingKey string

	// Exchange specifies the name of the message exchange in messaging systems like RabbitMQ.
	Exchange string

	// RequestBody contains the message body received by the consumer, typically representing the content of the request.
	RequestBody []byte
}

// Init initializes the Consumer by setting a default value for the Exchange field if it is empty.
func (c *Consumer) Init() {
	if c.Exchange == "" {
		c.Exchange = "default"
	}
}

// name constructs and returns a string combining the consumer's exchange, queue, and routing key, with defaults if unset.
func (c *Consumer) name() string {
	return c.Exchange + "/" + c.Queue + "/" + c.RoutingKey
}

// SetConsumer updates the transaction with consumer details such as name, request body, exchange, queue, and routing key.
func (txn *TxnRecord) SetConsumer(c *Consumer) {
	if nil == txn || nil == c {
		return
	}
	txn.name = c.name()
	internal.RequestBodyConsumerAttributes(txn.attrs, c.RequestBody)
	txn.attrs.Value().AddString(internal.AttributeConsumerExchange, c.Exchange)
	txn.attrs.Value().AddString(internal.AttributeConsumerQueue, c.Queue)
	txn.attrs.Value().AddString(internal.AttributeConsumerRoutingKey, c.RoutingKey)
}
