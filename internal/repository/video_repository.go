package repository

import (
	"github.com/alimarzban99/video-processor-service/internal/models"
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
