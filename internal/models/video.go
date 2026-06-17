package models

import "time"

type Video struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index;not null"`
	Title     string `gorm:"type:varchar(255);not null"`
	Path      string `gorm:"type:varchar(500);not null"`
	Status    string `gorm:"type:varchar(50);default:'pending'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
