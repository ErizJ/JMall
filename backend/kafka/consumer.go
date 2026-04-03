package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// ConsumeHandler processes a single Kafka message.
// Return nil to commit offset; return error to retry.
type ConsumeHandler func(ctx context.Context, key, value string) error

// ExhaustedHandler is called when all retries are exhausted for a message.
// Optional — used for cleanup like rolling back Redis state.
type ExhaustedHandler func(ctx context.Context, key, value string)

// Consumer wraps kafka-go Reader for consumer-group based consumption.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a Kafka consumer for the given topic and group.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       topic,
			GroupID:     groupID,
			MinBytes:    1,
			MaxBytes:    10e6, // 10MB
			StartOffset: kafka.LastOffset,
		}),
	}
}

// Start begins consuming messages in a blocking loop.
// Cancel ctx to stop. onExhausted is called when retries are exhausted (can be nil).
//
// 重试策略：同一条消息最多重试 3 次，退避 1s/2s，
// 全部失败则调用 onExhausted 做清理，然后提交 offset 跳过。
func (c *Consumer) Start(ctx context.Context, handler ConsumeHandler, onExhausted ExhaustedHandler) {
	const maxRetries = 3
	logx.Info("kafka consumer started")

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logx.Info("kafka consumer stopped")
				return
			}
			logx.Errorf("kafka fetch error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		key := string(msg.Key)
		val := string(msg.Value)

		var handleErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			handleErr = handler(ctx, key, val)
			if handleErr == nil {
				break
			}
			logx.Errorf("kafka consume error (attempt %d/%d): key=%s, err=%v",
				attempt, maxRetries, key, handleErr)
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
		}

		if handleErr != nil {
			logx.Errorf("kafka msg exhausted %d retries, skipping: key=%s", maxRetries, key)
			if onExhausted != nil {
				onExhausted(ctx, key, val)
			}
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			logx.Errorf("kafka commit error: %v", err)
		}
	}
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
