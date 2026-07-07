package repository

import (
	"context"
	"fmt"
	customerrors "github.com/alimarzban99/video-processor-service/pkg/errors"
	"gorm.io/gorm"
)

type GenericRepository[T any] struct {
	db *gorm.DB
}

func NewGenericRepository[T any](db *gorm.DB) *GenericRepository[T] {
	return &GenericRepository[T]{db: db}
}

func (r *GenericRepository[T]) Create(
	ctx context.Context,
	entity *T,
) error {

	err := r.db.
		WithContext(ctx).
		Create(entity).
		Error

	if err != nil {

		return &customerrors.DatabaseError{
			Query: "CREATE",
			Err:   err,
		}
	}

	return nil
}

func (r *GenericRepository[T]) FindByID(
	ctx context.Context,
	id uint,
) (*T, error) {

	var entity T

	err := r.db.
		WithContext(ctx).
		First(&entity, id).
		Error

	if err != nil {

		return nil, &customerrors.DatabaseError{
			Query: fmt.Sprintf("SELECT id=%d", id),
			Err:   err,
		}
	}

	return &entity, nil
}

func (r *GenericRepository[T]) Update(
	ctx context.Context,
	entity *T,
) error {

	err := r.db.
		WithContext(ctx).
		Save(entity).
		Error

	if err != nil {

		return &customerrors.DatabaseError{
			Query: "UPDATE",
			Err:   err,
		}
	}

	return nil
}

func (r *GenericRepository[T]) Delete(
	ctx context.Context,
	id uint,
) error {

	var entity T

	err := r.db.
		WithContext(ctx).
		Delete(&entity, id).
		Error

	if err != nil {

		return &customerrors.DatabaseError{
			Query: fmt.Sprintf("DELETE id=%d", id),
			Err:   err,
		}
	}

	return nil
}
