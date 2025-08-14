package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
)

const (
	kafkaBroker = "localhost:9092"
	topic       = "test-topic"
	consumerGroup = "test-consumer-group"
)

// Message represents a sample message structure
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
}

func main() {
	// Initialize logmanager application
	app := logmanager.NewApplication(
		logmanager.WithAppName("kafka-example"),
		logmanager.WithService("kafka"),
		logmanager.WithDebug(),
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// Create a wait group for goroutines
	var wg sync.WaitGroup

	// Start producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		runProducer(ctx, app)
	}()

	// Start consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		runConsumer(ctx, app)
	}()

	// Wait for termination signal
	<-sigterm
	fmt.Println("Shutting down Kafka example...")
	cancel()
	
	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("Kafka example stopped")
}

func runProducer(ctx context.Context, app *logmanager.Application) {
	// Configure Sarama
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	// Create producer
	producer, err := sarama.NewSyncProducer([]string{kafkaBroker}, config)
	if err != nil {
		fmt.Printf("Failed to create Kafka producer: %v\n", err)
		return
	}
	defer func() {
		if err := producer.Close(); err != nil {
			fmt.Printf("Failed to close producer: %v\n", err)
		}
	}()

	fmt.Printf("Kafka producer started - broker: %s, topic: %s\n", kafkaBroker, topic)

	// Produce messages
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	messageCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Producer shutting down - messages sent: %d\n", messageCount)
			return
		case <-ticker.C:
			messageCount++
			
			// Create message
			msg := Message{
				ID:        fmt.Sprintf("msg-%d", messageCount),
				Content:   fmt.Sprintf("Test message #%d", messageCount),
				Timestamp: time.Now(),
				UserID:    "user-123",
			}

			// Start a transaction for this message
			txn := app.Start("producer-msg-"+msg.ID, "kafka-producer", logmanager.TxnTypeOther)
			_ = txn.ToContext(ctx) // Context available if needed for logging
			
			// Marshal to JSON
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				txn.NoticeError(err)
				txn.End()
				continue
			}

			// Create Kafka message
			kafkaMsg := &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder(msg.ID),
				Value: sarama.ByteEncoder(msgBytes),
				Headers: []sarama.RecordHeader{
					{
						Key:   []byte("trace-id"),
						Value: []byte(txn.TraceID()),
					},
				},
			}

			// Send message
			partition, offset, err := producer.SendMessage(kafkaMsg)
			if err != nil {
				txn.NoticeError(err)
				// Transaction will log the error details at End()
				continue
			}

			// Transaction record is logged at End()
			txn.End()
			
			fmt.Printf("Message sent - ID: %s, partition: %d, offset: %d\n", msg.ID, partition, offset)
		}
	}
}

func runConsumer(ctx context.Context, app *logmanager.Application) {
	// Configure Sarama
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Create consumer group
	consumerGroupHandler := &ConsumerGroupHandler{
		app: app,
		ctx: ctx,
	}

	client, err := sarama.NewConsumerGroup([]string{kafkaBroker}, consumerGroup, config)
	if err != nil {
		fmt.Printf("Failed to create consumer group: %v\n", err)
		return
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("Failed to close consumer: %v\n", err)
		}
	}()

	fmt.Printf("Kafka consumer started - broker: %s, topic: %s, group: %s\n", kafkaBroker, topic, consumerGroup)

	// Consume messages
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Consumer shutting down")
			return
		default:
			// Consume messages from topic
			err := client.Consume(ctx, []string{topic}, consumerGroupHandler)
			if err != nil {
				fmt.Printf("Error consuming messages: %v\n", err)
				time.Sleep(5 * time.Second) // Wait before retrying
			}
		}
	}
}

// ConsumerGroupHandler represents a Sarama consumer group consumer
type ConsumerGroupHandler struct {
	app          *logmanager.Application
	ctx          context.Context
	messageCount int
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	fmt.Println("Consumer group session started")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	fmt.Printf("Consumer group session ended - messages processed: %d\n", h.messageCount)
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Process each message
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			h.messageCount++
			
			// Extract trace ID from headers (for tracking)
			var traceID string
			for _, header := range message.Headers {
				if string(header.Key) == "trace-id" {
					traceID = string(header.Value)
					break
				}
			}

			// Start a transaction for this message
			txnName := fmt.Sprintf("consumer-msg-%d", message.Offset)
			txn := h.app.Start(txnName, "kafka-consumer", logmanager.TxnTypeConsumer)
			
			// Transaction will use the trace ID from the message header
			_ = traceID // Track trace ID if needed
			_ = txn.ToContext(h.ctx) // Context available if needed for logging
			
			// Parse message
			var msg Message
			if err := json.Unmarshal(message.Value, &msg); err != nil {
				txn.NoticeError(err)
				// Error details are logged at End()
				session.MarkMessage(message, "")
				continue
			}
			
			fmt.Printf("Message received - ID: %s, offset: %d, partition: %d\n", msg.ID, message.Offset, message.Partition)

			// Simulate processing
			time.Sleep(100 * time.Millisecond)

			// Mark message as processed
			session.MarkMessage(message, "")

			// Transaction completed
			txn.End()

		case <-h.ctx.Done():
			return nil
		}
	}
}