package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// Producer wraps kafka-go Writer with JSON serialization.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a Kafka producer for the given topic.
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.Hash{},
			BatchSize:              100,
			BatchTimeout:           5 * time.Millisecond,
			RequiredAcks:           kafka.RequireOne,
			Compression:            kafka.Lz4,
			AllowAutoTopicCreation: true,
		},
	}
}

// Send serializes value to JSON and writes to Kafka.
func (p *Producer) Send(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
	})
}

// Close flushes and closes the producer.
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// MustNewProducer creates a producer and logs the topic.
func MustNewProducer(brokers []string, topic string) *Producer {
	logx.Infof("kafka producer starting: brokers=%v, topic=%s", brokers, topic)
	return NewProducer(brokers, topic)
}
