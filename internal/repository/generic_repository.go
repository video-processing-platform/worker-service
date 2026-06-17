package repository

import (
	"context"

	"gorm.io/gorm"
)

type GenericRepository[T any] struct {
	db *gorm.DB
}

func NewGenericRepository[T any](db *gorm.DB) *GenericRepository[T] {
	return &GenericRepository[T]{db: db}
}

func (r *GenericRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *GenericRepository[T]) FindByID(ctx context.Context, id uint) (*T, error) {
	var entity T

	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}

	return &entity, nil
}

func (r *GenericRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GenericRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}
