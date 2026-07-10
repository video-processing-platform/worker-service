package repository

import (
	"fmt"
	"github.com/alimarzban99/video-processor-service/internal/models"
	"github.com/alimarzban99/video-processor-service/pkg/database"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
)

type VideoRepository struct {
	*GenericRepository[models.Video]
}

func NewVideoRepository() *VideoRepository {
	db := database.DB()
	return &VideoRepository{
		GenericRepository: NewGenericRepository[models.Video](db),
	}
}

func (r *VideoRepository) UpdateStatus(
	videoID int,
	status int,
) error {

	result := r.db.
		Table("videos").
		Where("id = ?", videoID).
		Update("status", status)

	if result.Error != nil {
		return &customerrors.DatabaseError{
			Query: fmt.Sprintf(
				"UPDATE videos SET status=%d WHERE id=%d",
				status,
				videoID,
			),
			Err: result.Error,
		}
	}

	if result.RowsAffected == 0 {
		return customerrors.ErrNotFound
	}

	return nil
}
