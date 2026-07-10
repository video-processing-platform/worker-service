package workers

import (
	"context"
	"encoding/json"
	"github.com/alimarzban99/video-processor-service/config"
	"github.com/alimarzban99/video-processor-service/internal/models"
	"github.com/alimarzban99/video-processor-service/internal/services"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
	"github.com/alimarzban99/video-processor-service/pkg/logger"
	"github.com/alimarzban99/video-processor-service/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"sync"
)

type Consumer struct {
	ctx context.Context

	rmq          *rabbitmq.RabbitMQ
	videoService *services.VideoService
	workerCount  int

	jobs chan amqp.Delivery

	wg sync.WaitGroup
}

func NewConsumer(ctx context.Context, rmq *rabbitmq.RabbitMQ, videoService *services.VideoService) *Consumer {
	cfg := config.Cfg.Rabbit
	return &Consumer{
		ctx:          ctx,
		rmq:          rmq,
		videoService: videoService,
		workerCount:  cfg.WorkerCount,
		jobs:         make(chan amqp.Delivery),
	}
}

func (c *Consumer) Start() error {

	messages, err := c.rmq.Consume()
	if err != nil {
		return &customerrors.ProcessingError{
			Step: "start rabbitmq consumer",
			Err:  err,
		}
	}

	for i := 0; i < c.workerCount; i++ {

		go func(workerID int) {

			log.Printf("worker %d started", workerID)

			for msg := range c.jobs {

				c.wg.Add(1)

				func() {
					defer c.wg.Done()
					c.handle(msg)
				}()
			}

		}(i + 1)
	}

	for message := range messages {
		c.jobs <- message
	}

	log.Println("rabbitmq consumer stopped")

	return nil
}

func (c *Consumer) handle(msg amqp.Delivery) {

	var job models.VideoJob

	if err := json.Unmarshal(msg.Body, &job); err != nil {

		logger.Log.Error("invalid job payload", zap.Error(err), zap.ByteString("body", msg.Body))

		if ackErr := msg.Ack(false); ackErr != nil {
			logger.Log.Error("ack failed", zap.Error(ackErr))
		}

		return
	}

	log.Printf("received job: %s", string(msg.Body))
	err := c.videoService.Process(c.ctx, job)

	if err != nil {
		logger.Log.Error(
			"video processing failed",
			zap.Int("video_id", job.VideoID),
			zap.Error(err),
		)
	}

	if err == nil {
		if ackErr := msg.Ack(false); ackErr != nil {
			logger.Log.Error("ack failed", zap.Error(ackErr), zap.Int("video_id", job.VideoID))
		}
		return
	}

	retryCount := 1

	if v, ok := msg.Headers["x-retry-count"]; ok {
		switch t := v.(type) {
		case int32:
			retryCount = int(t)
		case int:
			retryCount = t
		}
	}

	log.Printf("retry count: %d", retryCount)

	if retryCount >= 3 {

		err := c.rmq.PublishDead(msg.Body, msg.Headers)

		if err != nil {

			logger.Log.Error("publish dead failed", zap.Error(err))

			if nackErr := msg.Nack(false, true); nackErr != nil {
				logger.Log.Error("nack failed", zap.Error(nackErr))
			}

			return
		}

		if err := msg.Ack(false); err != nil {
			logger.Log.Error("ack failed", zap.Error(err))
		}

		return
	}

	err = c.rmq.PublishRetry(msg.Body, retryCount+1)

	if err != nil {

		logger.Log.Error("publish retry failed", zap.Error(err))

		if nackErr := msg.Nack(false, true); nackErr != nil {
			logger.Log.Error("nack failed", zap.Error(nackErr))
		}

		return
	}

	if err := msg.Ack(false); err != nil {
		logger.Log.Error("ack failed", zap.Error(err))
	}
}

func (c *Consumer) Shutdown() {

	log.Println("stopping consumer...")

	if err := c.rmq.CancelConsumer(); err != nil {
		log.Printf("cancel consumer error: %v", err)
	}

	close(c.jobs)

	log.Println("waiting active jobs...")

	c.wg.Wait()

	log.Println("all workers finished")

	if err := c.rmq.Close(); err != nil {
		logger.Log.Error("rabbitmq close failed", zap.Error(err))
	}

	log.Println("shutdown completed")
}
