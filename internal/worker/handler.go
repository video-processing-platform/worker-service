package worker

import (
	"fmt"
	"log"
)

type Publisher interface {
	Publish(exchange, key string, body []byte, headers map[string]any) error
}

func VideoHandler(publisher Publisher) func(Job) {

	return func(job Job) {

		err := processVideo(job.Body)

		if err != nil {

			if job.RetryCount < 3 {

				publisher.Publish(
					"video.exchange",
					"video.retry",
					job.Body,
					map[string]any{
						"retry_count": job.RetryCount + 1,
					},
				)

			} else {

				publisher.Publish(
					"video.exchange",
					"video.dead",
					job.Body,
					nil,
				)
			}

			return
		}

		log.Println("SUCCESS")
	}
}

func processVideo(body []byte) error {
	log.Println("Processing video:", string(body))

	return fmt.Errorf("fake error")

}
