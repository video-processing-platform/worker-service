package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	conn, _ := amqp.Dial(
		"amqp://guest:guest@localhost:5672/",
	)

	ch, _ := conn.Channel()

	ch.Publish(
		"",
		"video.processing.queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body: []byte(`{
				"video_id":16,
				"user_id":7,
				"object_key":"7/2026-07-10-16-31-21/main/file_example_MP4_640_3MG.mp4"
			}`),
		},
	)
}
