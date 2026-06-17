package rabbitmq

import (
	"fmt"
	"github.com/alimarzban99/video-processor-service/internal/worker"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Client struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func Init() {

	rabbit, err := new()

	if err != nil {
		log.Fatal(err)
	}

	defer rabbit.Conn.Close()
	defer rabbit.Channel.Close()
	err = rabbit.Setup()

	if err != nil {
		log.Fatal(err)
	}

	pool := worker.NewPool(1)

	pool.Start(worker.VideoHandler(rabbit))

	err = rabbit.Consume(pool)

	if err != nil {
		log.Fatal(err)
	}

	select {}
}

func new() (*Client, error) {

	rabbitURL := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		"guest",
		"guest",
		"localhost",
		"5672",
	)
	conn, err := amqp.Dial(rabbitURL)

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn:    conn,
		Channel: ch,
	}, nil
}
