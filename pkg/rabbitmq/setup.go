package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

func (c *Client) Setup() error {

	err := c.Channel.ExchangeDeclare(
		"video.exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.Channel.QueueDeclare(
		"video.processing.queue",
		true,
		false,
		false,
		false,
		nil,
	)

	retryArgs := amqp.Table{
		"x-message-ttl":             int32(30000),
		"x-dead-letter-exchange":    "video.exchange",
		"x-dead-letter-routing-key": "video.process",
	}

	_, err = c.Channel.QueueDeclare(
		"video.retry.queue",
		true,
		false,
		false,
		false,
		retryArgs,
	)

	err = c.Channel.QueueBind(
		"video.retry.queue",
		"video.retry",
		"video.exchange",
		false,
		nil,
	)

	_, err = c.Channel.QueueDeclare(
		"video.dead.queue",
		true,
		false,
		false,
		false,
		nil,
	)

	err = c.Channel.QueueBind(
		"video.dead.queue",
		"video.dead",
		"video.exchange",
		false,
		nil,
	)

	if err != nil {
		return err
	}

	err = c.Channel.QueueBind(
		"video.processing.queue",
		"video.process",
		"video.exchange",
		false,
		nil,
	)

	return err
}
