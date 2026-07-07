package models

import (
	"gorm.io/gorm"
	"time"
)

type Video struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index;not null"`
	Title     string `gorm:"type:varchar(255);not null"`
	Path      string `gorm:"type:varchar(500);not null"`
	Status    uint   `gorm:"type:s(50);default:'pending'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
