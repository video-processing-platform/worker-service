package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

func (c *Client) Publish(exchange, key string, body []byte, headers map[string]any) error {

	return c.Channel.Publish(
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Headers:     amqp.Table(headers),
			Body:        body,
		},
	)
}
