package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	VideosProcessedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "videos_processed_total",
			Help: "Total successfully processed videos",
		},
	)

	VideosFailedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "videos_failed_total",
			Help: "Total failed videos",
		},
	)

	ProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "processing_duration_seconds",
			Help:    "Video processing duration",
			Buckets: prometheus.DefBuckets,
		},
	)

	RabbitMQMessagesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_total",
			Help: "Total RabbitMQ consumed messages",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		VideosProcessedTotal,
		VideosFailedTotal,
		ProcessingDuration,
		RabbitMQMessagesTotal,
	)
}
