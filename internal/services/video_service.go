package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/alimarzban99/video-processor-service/internal/enums"
	grpcclient "github.com/alimarzban99/video-processor-service/internal/grpc"
	"github.com/alimarzban99/video-processor-service/internal/models"
	"github.com/alimarzban99/video-processor-service/internal/repository"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
	"github.com/alimarzban99/video-processor-service/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"time"
)

type VideoService struct {
	ffmpeg     *FFmpegService
	videoRepo  *repository.VideoRepository
	sem        *semaphore.Weighted
	jobTimeout time.Duration
}

func NewVideoService(
	ffmpeg *FFmpegService,
	videoRepo *repository.VideoRepository,
	maxFFmpeg int,
	jobTimeout time.Duration,

) *VideoService {

	return &VideoService{
		ffmpeg:     ffmpeg,
		videoRepo:  videoRepo,
		sem:        semaphore.NewWeighted(int64(maxFFmpeg)),
		jobTimeout: jobTimeout,
	}
}

func (s *VideoService) Process(
	ctx context.Context,
	job models.VideoJob,
) error {

	ctx, cancel := context.WithTimeout(
		ctx,
		s.jobTimeout,
	)

	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	if err := s.videoRepo.UpdateStatus(
		job.VideoID,
		int(enums.Processing),
	); err != nil {

		return &customerrors.StorageError{
			Operation: "update video status to processing",
			Err:       err,
		}
	}

	notificationClient, _ := grpcclient.NewNotificationClient()

	_ = notificationClient.Send(
		"1",
		"Processing Started",
		"Your video processing has started",
	)

	g.Go(func() error {
		return s.runFFmpeg(
			ctx,
			func() error {
				return s.ffmpeg.Convert1080(
					ctx,
					job.File,
					job.VideoID,
				)

			},
		)

	})

	g.Go(func() error {
		return s.runFFmpeg(
			ctx,
			func() error {
				return s.ffmpeg.Convert720(
					ctx,
					job.File,
					job.VideoID,
				)

			},
		)

	})

	g.Go(func() error {
		return s.runFFmpeg(
			ctx,
			func() error {
				return s.ffmpeg.Convert480(
					ctx,
					job.File,
					job.VideoID,
				)

			},
		)

	})

	g.Go(func() error {
		return s.runFFmpeg(
			ctx,
			func() error {
				return s.ffmpeg.Thumbnail(
					ctx,
					job.File,
					job.VideoID,
				)

			},
		)

	})

	err := g.Wait()

	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {

			return &customerrors.ProcessingError{
				Step: "video processing timeout",
				Err: fmt.Errorf(
					"deadline exceeded: %w",
					err,
				),
			}
		}

		if err := s.videoRepo.UpdateStatus(
			job.VideoID,
			int(enums.Failed),
		); err != nil {

			logger.Log.Error(
				"failed to update video status",
				zap.Int("video_id", job.VideoID),
				zap.Int("status", int(enums.Failed)),
				zap.Error(err),
			)
		}

		_ = notificationClient.Send(
			"1",
			"Processing Failed",
			"Your video is ready",
		)

		return &customerrors.ProcessingError{
			Step: "ffmpeg processing",
			Err:  err,
		}
	}
	if err := s.videoRepo.UpdateStatus(
		job.VideoID,
		int(enums.Completed),
	); err != nil {

		return &customerrors.StorageError{
			Operation: "update video status to completed",
			Err:       err,
		}
	}

	_ = notificationClient.Send(
		"1",
		"Processing completed",
		"Your video is ready",
	)

	return nil
}

func (s *VideoService) runFFmpeg(
	ctx context.Context,
	fn func() error,
) error {

	if err := s.sem.Acquire(ctx, 1); err != nil {

		return &customerrors.ProcessingError{
			Step: "acquire ffmpeg semaphore",
			Err: fmt.Errorf(
				"failed to acquire semaphore: %w",
				err,
			),
		}
	}

	defer s.sem.Release(1)

	if err := fn(); err != nil {

		return &customerrors.ProcessingError{
			Step: "execute ffmpeg job",
			Err: fmt.Errorf(
				"ffmpeg execution failed: %w",
				err,
			),
		}
	}

	return nil
}
