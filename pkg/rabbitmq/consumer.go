package rabbitmq

import (
	"github.com/alimarzban99/video-processor-service/internal/worker"
	_ "github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (c *Client) Consume(pool *worker.Pool) error {

	msgs, err := c.Channel.Consume(
		"video.processing.queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {

			retryCount := getRetryCount(msg)

			job := worker.Job{
				Body:       msg.Body,
				RetryCount: retryCount,
			}

			// ارسال به worker pool
			pool.Submit(job)

			msg.Ack(false)
		}
	}()

	return nil
}

func getRetryCount(msg amqp.Delivery) int {

	if msg.Headers == nil {
		return 0
	}

	if v, ok := msg.Headers["retry_count"]; ok {

		switch val := v.(type) {
		case int32:
			return int(val)
		case int64:
			return int(val)
		case int:
			return val
		}
	}

	return 0
}
