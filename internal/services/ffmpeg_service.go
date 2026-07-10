package services

import (
	"context"
	"fmt"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
	"os"
	"os/exec"
)

type FFmpegService struct{}

func NewFFmpegService() *FFmpegService {
	return &FFmpegService{}
}

func (s *FFmpegService) Convert(ctx context.Context, input string, output string, quality int, videoID int) error {

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", input,
		"-vf", fmt.Sprintf("scale=-2:%d", quality),
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-c:a", "aac",
		output,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return s.wrapError(fmt.Sprintf("convert %dp", quality), videoID, err)
	}

	return nil
}

func (s *FFmpegService) Thumbnail(ctx context.Context, input string, output string, videoID int) error {

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", input,
		"-ss", "00:00:05",
		"-vframes", "1",
		output,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {

		return s.wrapError("generate thumbnail", videoID, err)
	}

	return nil
}
func (s *FFmpegService) wrapError(step string, videoID int, err error) error {

	if err == nil {
		return nil
	}

	return &customerrors.ProcessingError{
		Step: step,
		Err:  fmt.Errorf("video %d: %w", videoID, err),
	}
}
