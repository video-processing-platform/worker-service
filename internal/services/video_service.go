package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/alimarzban99/video-processor-service/config"
	"github.com/alimarzban99/video-processor-service/internal/enums"
	grpcclient "github.com/alimarzban99/video-processor-service/internal/grpc"
	"github.com/alimarzban99/video-processor-service/internal/models"
	"github.com/alimarzban99/video-processor-service/internal/repository"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
	"github.com/alimarzban99/video-processor-service/pkg/logger"
	"github.com/alimarzban99/video-processor-service/pkg/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type VideoService struct {
	storage      *storage.MinioStorage
	ffmpeg       *FFmpegService
	videoRepo    *repository.VideoRepository
	notification *grpcclient.NotificationClient
	sem          *semaphore.Weighted
	jobTimeout   time.Duration
}

func NewVideoService(
	storage *storage.MinioStorage,
	ffmpeg *FFmpegService,
	videoRepo *repository.VideoRepository,
	notification *grpcclient.NotificationClient) *VideoService {

	cfg := config.Cfg.Rabbit

	return &VideoService{
		storage:      storage,
		ffmpeg:       ffmpeg,
		videoRepo:    videoRepo,
		notification: notification,
		sem:          semaphore.NewWeighted(int64(cfg.MaxFFmpegWorker)),
		jobTimeout:   time.Duration(cfg.JobTimeout) * time.Second,
	}
}

func (s *VideoService) Process(ctx context.Context, job models.VideoJob) error {

	logger.Log.Info(
		"job timeout",
		zap.Duration("timeout", s.jobTimeout),
	)
	jobCtx, cancel := context.WithTimeout(ctx, s.jobTimeout)
	defer cancel()

	if err := s.videoRepo.UpdateStatus(job.VideoID, int(enums.Processing)); err != nil {
		return &customerrors.StorageError{
			Operation: "update video status to processing",
			Err:       err,
		}
	}

	_ = s.notification.Send(strconv.Itoa(job.VideoID), "Processing Started", "Your video processing has started")

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("video-%d-*", job.VideoID))
	if err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	input := filepath.Join(tempDir, "input.mp4")

	if err := s.storage.DownloadFile(jobCtx, job.ObjectKey, input); err != nil {
		return fmt.Errorf("download input file: %w", err)
	}

	outputs := map[string]string{
		"1080":      filepath.Join(tempDir, "1080.mp4"),
		"720":       filepath.Join(tempDir, "720.mp4"),
		"480":       filepath.Join(tempDir, "480.mp4"),
		"thumbnail": filepath.Join(tempDir, "thumbnail.jpg"),
	}

	var resolutions = []int{1080, 720, 480}
	ffmpegGroup, ffmpegCtx := errgroup.WithContext(jobCtx)

	for _, resolution := range resolutions {

		resolution := resolution

		ffmpegGroup.Go(func() error {

			output := outputs[strconv.Itoa(resolution)]

			return s.runFFmpeg(
				ffmpegCtx,
				func() error {
					return s.ffmpeg.Convert(ffmpegCtx, input, output, resolution, job.VideoID)
				},
			)

		})
	}

	ffmpegGroup.Go(func() error {

		return s.runFFmpeg(
			ffmpegCtx,
			func() error {
				return s.ffmpeg.Thumbnail(ffmpegCtx, input, outputs["thumbnail"], job.VideoID)
			},
		)

	})

	if err := ffmpegGroup.Wait(); err != nil {

		if errors.Is(err, context.DeadlineExceeded) {

			return &customerrors.ProcessingError{
				Step: "video processing timeout",
				Err:  fmt.Errorf("deadline exceeded: %w", err),
			}
		}

		if updateErr := s.videoRepo.UpdateStatus(job.VideoID, int(enums.Failed)); updateErr != nil {

			logger.Log.Error("failed to update video status", zap.Int("video_id", job.VideoID), zap.Error(updateErr))

		}

		_ = s.notification.Send(strconv.Itoa(job.VideoID), "Processing Failed", "Video processing failed")

		return &customerrors.ProcessingError{
			Step: "ffmpeg processing",
			Err:  err,
		}
	}

	if err := ctx.Err(); err != nil {
		logger.Log.Error(
			"ctx already canceled",
			zap.Error(err),
		)
	}
	uploadGroup, uploadCtx := errgroup.WithContext(jobCtx)

	for variant, localFile := range outputs {

		variant := variant
		localFile := localFile

		uploadGroup.Go(func() error {

			objectKey, err := s.buildOutputKey(job.ObjectKey, variant)

			if err != nil {
				return err
			}

			logger.Log.Info(
				"uploading file",
				zap.String("variant", variant),
				zap.String("local_file", localFile),
				zap.String("object_key", objectKey),
			)

			return s.storage.UploadFile(uploadCtx, objectKey, localFile)

		})

	}

	if err := uploadGroup.Wait(); err != nil {

		if updateErr := s.videoRepo.UpdateStatus(job.VideoID, int(enums.Failed)); updateErr != nil {
			logger.Log.Error("failed to update video status", zap.Int("video_id", job.VideoID), zap.Error(updateErr))
		}

		return fmt.Errorf("upload outputs: %w", err)
	}

	if err := s.videoRepo.UpdateStatus(job.VideoID, int(enums.Completed)); err != nil {

		return &customerrors.StorageError{
			Operation: "update video status to completed",
			Err:       err,
		}
	}

	_ = s.notification.Send(strconv.Itoa(job.VideoID), "Processing Completed", "Your video is ready")

	return nil
}

func (s *VideoService) runFFmpeg(ctx context.Context, fn func() error) error {

	if err := s.sem.Acquire(ctx, 1); err != nil {

		return &customerrors.ProcessingError{
			Step: "acquire ffmpeg semaphore",
			Err:  fmt.Errorf("failed to acquire semaphore: %w", err),
		}
	}

	defer s.sem.Release(1)

	if err := fn(); err != nil {

		return &customerrors.ProcessingError{
			Step: "execute ffmpeg job",
			Err:  fmt.Errorf("ffmpeg execution failed: %w", err),
		}
	}

	return nil
}

func (s *VideoService) buildOutputKey(objectKey string, variant string) (string, error) {

	dir := path.Dir(objectKey)
	// 7/2026-07-10-14-41-20/main

	base := path.Base(objectKey)
	// file_example_MP4_640_3MG.mp4

	ext := path.Ext(base)
	// .mp4

	name := strings.TrimSuffix(base, ext)
	// file_example_MP4_640_3MG

	parts := strings.Split(dir, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid object key: %s", objectKey)
	}

	userID := parts[0]
	dateTime := parts[1]

	if variant == "thumbnail" {
		ext = ".jpg"
	}

	return fmt.Sprintf("%s/%s/converted/%s_%s%s", userID, dateTime, name, variant, ext), nil
}
