package services

import (
	"context"
	"fmt"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
	"os/exec"
)

type FFmpegService struct{}

func NewFFmpegService() *FFmpegService {
	return &FFmpegService{}
}

func (s *FFmpegService) Convert720(
	ctx context.Context,
	input string,
	videoID int,
) error {

	output := fmt.Sprintf(
		"output/%d_720.mp4",
		videoID,
	)

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-i",
		input,
		"-vf",
		"scale=-2:720",
		output,
	)

	if err := cmd.Run(); err != nil {

		return wrapError(
			"ffmpeg convert 720p",
			videoID,
			err,
		)
	}

	return nil
}

func (s *FFmpegService) Convert480(
	ctx context.Context,
	input string,
	videoID int,
) error {

	output := fmt.Sprintf(
		"output/%d_480.mp4",
		videoID,
	)

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-i",
		input,
		"-vf",
		"scale=-2:480",
		output,
	)

	if err := cmd.Run(); err != nil {

		return wrapError(
			"ffmpeg convert 480p",
			videoID,
			err,
		)
	}

	return nil
}

func (s *FFmpegService) Convert1080(
	ctx context.Context,
	input string,
	videoID int,
) error {

	output := fmt.Sprintf(
		"output/%d_1080.mp4",
		videoID,
	)

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-i",
		input,
		"-vf",
		"scale=-2:1080",
		output,
	)

	if err := cmd.Run(); err != nil {
		return wrapError(
			"ffmpeg convert 1080p",
			videoID,
			err,
		)
	}

	return nil
}

func (s *FFmpegService) Thumbnail(
	ctx context.Context,
	input string,
	videoID int,
) error {

	output := fmt.Sprintf(
		"output/%d_thumb.jpg",
		videoID,
	)

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-i",
		input,
		"-ss",
		"00:00:05",
		"-vframes",
		"1",
		output,
	)

	if err := cmd.Run(); err != nil {
		return wrapError(
			"ffmpeg thumbnail generation",
			videoID,
			err,
		)
	}

	return nil
}

func wrapError(
	step string,
	videoID int,
	err error,
) error {

	if err == nil {
		return nil
	}

	return &customerrors.ProcessingError{
		Step: step,
		Err: fmt.Errorf(
			"video %d: %w",
			videoID,
			err,
		),
	}
}
