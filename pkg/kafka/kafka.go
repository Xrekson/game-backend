package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
)

type Producer struct {
	writers map[string]*kafka.Writer
	brokers []string
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	return &Producer{
		writers: make(map[string]*kafka.Writer),
		brokers: cfg.Brokers,
	}
}

func (p *Producer) getWriter(topic string) *kafka.Writer {
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}
	p.writers[topic] = w
	return w
}

func (p *Producer) Publish(ctx context.Context, topic, key string, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	writer := p.getWriter(topic)
	msg := kafka.Message{
		Key:   []byte(key),
		Value: bytes,
		Time:  time.Now(),
	}

	err = writer.WriteMessages(ctx, msg)
	if err != nil {
		logger.Error("Failed to publish message to topic %s: %v", topic, err)
		return err
	}

	logger.Debug("Published message to Kafka topic %s key %s", topic, key)
	return nil
}

func (p *Producer) Close() {
	for _, w := range p.writers {
		_ = w.Close()
	}
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(cfg config.KafkaConfig, topic string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    topic,
		MinBytes: 10,
		MaxBytes: 10 * 1024 * 1024,
	})

	return &Consumer{reader: reader}
}

func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
