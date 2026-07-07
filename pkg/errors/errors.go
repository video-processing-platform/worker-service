package errors

import (
	"errors"
	"fmt"
)

var (
	ErrPublishNotConfirmed = errors.New("publish not confirmed")
	ErrRabbitMQConnection  = errors.New("rabbitmq connection failed")
	ErrRabbitMQChannel     = errors.New("rabbitmq channel failed")
	ErrNotFound            = errors.New("resource not found")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: field=%s message=%s",
		e.Field,
		e.Message,
	)
}

type StorageError struct {
	Operation string
	Err       error
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage error during %s: %v",
		e.Operation,
		e.Err,
	)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

type ProcessingError struct {
	Step string
	Err  error
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("processing error at step %s: %v",
		e.Step,
		e.Err,
	)
}

func (e *ProcessingError) Unwrap() error {
	return e.Err
}

type DatabaseError struct {
	Query string
	Err   error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error executing query [%s]: %v",
		e.Query,
		e.Err,
	)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}
