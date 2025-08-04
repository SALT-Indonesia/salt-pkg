# Event Manager

An event manager who publishes and subscribes to events using RabbitMQ.

## Publish

In your microservice, you need to define these environment variables:

| Variable     | Example                  | Description   |
|--------------|--------------------------|---------------|
| RABBITMQ_URL | `amqp://@localhost:5672` | RabbitMQ url. |

### Usage

```go
type message struct { // your struct for your messages
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

myMessage := &message{
    Title:       "This is my first message",
    Description: "This is the description of my first message.",
}

// you can put it on your infrastructure layers
if err := eventmanager.Publish(context.Background(), "mytopic", myMessage); err != nil {
    log.Fatal(err)
}
```

## Subscribe

In your microservice, you need to define these environment variables:

| Variable                 | Example                  | Description                                                                                            |
|--------------------------|--------------------------|--------------------------------------------------------------------------------------------------------|
| RABBITMQ_CHANNELS        | `1`                      | RabbitMQ total channels. Default is 1. Maximum is 10.                                                  |
| RABBITMQ_RECONNECT_DELAY | `2s`                     | RabbitMQ delay duration for reconnect. Default is 2 seconds.                                           |
| RABBITMQ_PREFETCH_COUNT  | `10`                     | RabbitMQ prefetch count. Default is 10.                                                                |
| RABBITMQ_REFRESH_DELAY   | `1h`                     | RabbitMQ delay duration for refreshing counter. Default is 1 hour.                                     |
| RABBITMQ_MAX_RETRY       | `5`                      | RabbitMQ maximum retry for intrastructure errors. Default is 5.                                        |
| RABBITMQ_RETRY_DELAY     | `1m`                     | RabbitMQ delay duration for retrying fail messages due to infrastructure errors. Default is 1 minutes. |
| RABBITMQ_URL             | `amqp://@localhost:5672` | RabbitMQ url.                                                                                          |

### Usage

```go
type message struct { // your struct your your incoming messages
	Title       string `json:"title"`
	Description string `json:"description"`
}

func handle(m message) error { // an example for handlers
	log.Println(m)

	return nil
}

// you need to run it directly on your main
func main() {
	eventmanager.Subscribe(context.Background(), "myservice", "mytopic", []eventmanager.Handler[message]{handle})
}
```
