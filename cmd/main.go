package main

import (
	"context"
	"github.com/alimarzban99/video-processor-service/config"
	"github.com/alimarzban99/video-processor-service/internal/repository"
	"github.com/alimarzban99/video-processor-service/internal/services"
	"github.com/alimarzban99/video-processor-service/internal/workers"
	"github.com/alimarzban99/video-processor-service/pkg/cache"
	"github.com/alimarzban99/video-processor-service/pkg/database"
	"github.com/alimarzban99/video-processor-service/pkg/logger"
	"github.com/alimarzban99/video-processor-service/pkg/metrics"
	"github.com/alimarzban99/video-processor-service/pkg/rabbitmq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	if err := logger.Init(); err != nil {
		log.Fatal(err)
	}

	defer logger.Log.Sync()

	config.Load()

	metrics.Register()
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		logger.Log.Info("metrics server started", zap.String("addr", ":2112"))

		if err := http.ListenAndServe(":2112", nil); err != nil {
			logger.Log.Error("metrics server failed", zap.Error(err))
		}
	}()

	if err := database.Init(); err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := cache.Init(); err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	rmq, err := rabbitmq.New()
	if err != nil {
		log.Fatal(err)
	}

	ffmpeg := services.NewFFmpegService()

	cfg := config.Cfg.Rabbit

	db := database.DB()

	videoRepo := repository.NewVideoRepository(db)
	videoService := services.NewVideoService(ffmpeg, videoRepo, cfg.MaxFFmpegWorker, cfg.JobTimeout)

	ctx, cancel := context.WithCancel(context.Background())

	consumer := workers.NewConsumer(
		ctx,
		rmq,
		videoService,
		cfg.WorkerCount,
	)

	go func() {

		if err := consumer.Start(); err != nil {
			log.Printf(
				"consumer stopped: %v",
				err,
			)
		}

	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	sig := <-signalChan

	log.Printf(
		"received signal: %s",
		sig.String(),
	)

	cancel()
	consumer.Shutdown()

	log.Println(
		"application stopped",
	)
}
