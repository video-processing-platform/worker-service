package repository

import (
	"fmt"
	"github.com/alimarzban99/video-processor-service/internal/models"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
	"gorm.io/gorm"
)

type VideoRepository struct {
	*GenericRepository[models.Video]
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
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
