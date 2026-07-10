package main

import (
	"context"
	"github.com/alimarzban99/video-processor-service/config"
	grpcclient "github.com/alimarzban99/video-processor-service/internal/grpc"
	"github.com/alimarzban99/video-processor-service/internal/repository"
	"github.com/alimarzban99/video-processor-service/internal/services"
	"github.com/alimarzban99/video-processor-service/internal/workers"
	"github.com/alimarzban99/video-processor-service/pkg/cache"
	"github.com/alimarzban99/video-processor-service/pkg/database"
	"github.com/alimarzban99/video-processor-service/pkg/logger"
	"github.com/alimarzban99/video-processor-service/pkg/metrics"
	"github.com/alimarzban99/video-processor-service/pkg/rabbitmq"
	"github.com/alimarzban99/video-processor-service/pkg/storage"
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

	minioStorage, err := storage.NewMinio()
	if err != nil {
		log.Fatal(err)
	}

	rabbitMQ, err := rabbitmq.New()
	if err != nil {
		log.Fatal(err)
	}

	ffmpeg := services.NewFFmpegService()
	videoRepo := repository.NewVideoRepository()
	notification, _ := grpcclient.NewNotificationClient()
	videoService := services.NewVideoService(minioStorage, ffmpeg, videoRepo, notification)

	ctx, cancel := context.WithCancel(context.Background())

	consumer := workers.NewConsumer(ctx, rabbitMQ, videoService)

	go func() {

		if err := consumer.Start(); err != nil {
			log.Printf("consumer stopped: %v", err)
		}

	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signalChan

	log.Printf("received signal: %s", sig.String())

	cancel()
	consumer.Shutdown()

	log.Println("application stopped")
}
