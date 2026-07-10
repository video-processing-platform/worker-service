package rabbitmq

import (
	"fmt"
	"github.com/alimarzban99/video-processor-service/config"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ProcessingQueue = "video.processing.queue"
	RetryQueue      = "video.retry.queue"
	DeadQueue       = "video.dead.queue"
	ConsumerTag     = "video-consumer"
)

type RabbitMQ struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

func New() (*RabbitMQ, error) {

	cfg := config.Cfg.Rabbit
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, &customerrors.StorageError{
			Operation: "rabbitmq connect",
			Err:       fmt.Errorf("%w: %w", customerrors.ErrRabbitMQConnection, err),
		}
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, &customerrors.StorageError{
			Operation: "rabbitmq create channel",
			Err:       fmt.Errorf("%w: %w", customerrors.ErrRabbitMQChannel, err),
		}
	}

	err = ch.Confirm(false)
	if err != nil {
		return nil, err
	}

	err = ch.Qos(5, 0, false)

	if err != nil {
		return nil, err
	}

	if err := declareQueues(ch); err != nil {
		return nil, err
	}

	return &RabbitMQ{
		Conn: conn,
		Ch:   ch,
	}, nil
}

func declareQueues(ch *amqp.Channel) error {

	_, err := ch.QueueDeclare(
		ProcessingQueue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {

		return &customerrors.StorageError{
			Operation: "declare processing queue",
			Err:       err,
		}
	}

	retryArgs := amqp.Table{
		"x-message-ttl":             30000,
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": ProcessingQueue,
	}

	_, err = ch.QueueDeclare(
		RetryQueue,
		true,
		false,
		false,
		false,
		retryArgs,
	)

	if err != nil {
		return &customerrors.StorageError{
			Operation: "declare retry queue",
			Err:       err,
		}
	}

	_, err = ch.QueueDeclare(
		DeadQueue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {

		return &customerrors.StorageError{
			Operation: "declare dead queue",
			Err:       err,
		}
	}

	return nil
}

func (r *RabbitMQ) Consume() (<-chan amqp.Delivery, error) {

	return r.Ch.Consume(
		ProcessingQueue,
		ConsumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) PublishRetry(body []byte, retryCount int) error {

	confirms := r.Ch.NotifyPublish(
		make(chan amqp.Confirmation, 1),
	)

	err := r.Ch.Publish(
		"",
		RetryQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers: amqp.Table{
				"x-retry-count": retryCount,
			},
		},
	)

	if err != nil {

		return &customerrors.StorageError{
			Operation: "publish retry message",
			Err:       err,
		}
	}

	confirm := <-confirms

	if !confirm.Ack {

		return &customerrors.StorageError{
			Operation: "confirm retry publish",
			Err:       customerrors.ErrPublishNotConfirmed,
		}
	}

	return nil
}

func (r *RabbitMQ) PublishDead(body []byte, headers amqp.Table) error {

	confirms := r.Ch.NotifyPublish(
		make(chan amqp.Confirmation, 1),
	)

	err := r.Ch.Publish(
		"",
		DeadQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers:      headers,
		},
	)

	if err != nil {

		return &customerrors.StorageError{
			Operation: "publish dead letter",
			Err:       err,
		}
	}

	confirm := <-confirms

	if !confirm.Ack {

		return &customerrors.StorageError{
			Operation: "confirm dead publish",
			Err:       customerrors.ErrPublishNotConfirmed,
		}
	}

	return nil
}

func (r *RabbitMQ) Close() error {

	if err := r.Ch.Close(); err != nil {

		return &customerrors.StorageError{
			Operation: "close rabbitmq channel",
			Err:       err,
		}
	}

	if err := r.Conn.Close(); err != nil {

		return &customerrors.StorageError{
			Operation: "close rabbitmq connection",
			Err:       err,
		}
	}

	return nil
}

func (r *RabbitMQ) CancelConsumer() error {

	if err := r.Ch.Cancel(
		ConsumerTag,
		false,
	); err != nil {

		return &customerrors.StorageError{
			Operation: "cancel consumer",
			Err:       err,
		}
	}

	return nil
}
