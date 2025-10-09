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
	kafkaBroker   = "localhost:9092"
	topic         = "test-topic"
	consumerGroup = "test-consumer-group"
)

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
}

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("kafka-example"),
		logmanager.WithService("kafka"),
		logmanager.WithDebug(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupGracefulShutdown(cancel)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		runProducer(ctx, app)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runConsumer(ctx, app)
	}()

	wg.Wait()
	fmt.Println("Kafka example stopped")
}

func setupGracefulShutdown(cancel context.CancelFunc) {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigterm
		fmt.Println("Shutting down Kafka example...")
		cancel()
	}()
}

func runProducer(ctx context.Context, app *logmanager.Application) {
	producer, err := createProducer()
	if err != nil {
		fmt.Printf("Failed to create producer: %v\n", err)
		return
	}
	defer producer.Close()

	fmt.Printf("Kafka producer started - broker: %s, topic: %s\n", kafkaBroker, topic)

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
			sendMessage(app, producer, messageCount)
		}
	}
}

func createProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	return sarama.NewSyncProducer([]string{kafkaBroker}, config)
}

func sendMessage(app *logmanager.Application, producer sarama.SyncProducer, count int) {
	msg := Message{
		ID:        fmt.Sprintf("msg-%d", count),
		Content:   fmt.Sprintf("Test message #%d", count),
		Timestamp: time.Now(),
		UserID:    "user-123",
	}

	txn := app.Start("producer-msg-"+msg.ID, "kafka-producer", logmanager.TxnTypeOther)
	defer txn.End()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		txn.NoticeError(err)
		return
	}

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

	partition, offset, err := producer.SendMessage(kafkaMsg)
	if err != nil {
		txn.NoticeError(err)
		return
	}

	fmt.Printf("Message sent - ID: %s, partition: %d, offset: %d\n", msg.ID, partition, offset)
}

func runConsumer(ctx context.Context, app *logmanager.Application) {
	handler := &ConsumerHandler{app: app, ctx: ctx}

	client, err := sarama.NewConsumerGroup([]string{kafkaBroker}, consumerGroup, createConsumerConfig())
	if err != nil {
		fmt.Printf("Failed to create consumer group: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("Kafka consumer started - broker: %s, topic: %s, group: %s\n", kafkaBroker, topic, consumerGroup)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Consumer shutting down")
			return
		default:
			if err := client.Consume(ctx, []string{topic}, handler); err != nil {
				fmt.Printf("Error consuming messages: %v\n", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func createConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	return config
}

type ConsumerHandler struct {
	app          *logmanager.Application
	ctx          context.Context
	messageCount int
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	fmt.Println("Consumer group session started")
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	fmt.Printf("Consumer group session ended - messages processed: %d\n", h.messageCount)
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}
			h.processMessage(session, message)
		case <-h.ctx.Done():
			return nil
		}
	}
}

func (h *ConsumerHandler) processMessage(session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) {
	h.messageCount++

	txnName := fmt.Sprintf("consumer-msg-%d", message.Offset)
	txn := h.app.Start(txnName, "kafka-consumer", logmanager.TxnTypeConsumer)
	defer txn.End()

	var msg Message
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		txn.NoticeError(err)
		session.MarkMessage(message, "")
		return
	}

	fmt.Printf("Message received - ID: %s, offset: %d, partition: %d\n", msg.ID, message.Offset, message.Partition)

	time.Sleep(100 * time.Millisecond)
	session.MarkMessage(message, "")
}