package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseErrorBuilder struct {
}

func NewDatabaseErrorBuilder() *DatabaseErrorBuilder {
	return &DatabaseErrorBuilder{}
}

func (that *DatabaseErrorBuilder) Build(ctx context.Context, err error, msg string) error {
	if err == nil {
		return nil
	}

	var pg *pgconn.PgError
	if !errors.As(err, &pg) {
		return err
	}

	de := &DomainError{
		Code:    pg.Code,
		Message: pg.Message,
		Detail:  pg.Detail,
		Hint:    pg.Hint,
	}

	switch pg.Code {
	case "23505":
		return fmt.Errorf("%w: %w", ErrAlreadyExists, de)
	case "23503", "23514", "23502", "22P02", "22023":
		return fmt.Errorf("%w: %w", ErrInvalid, de)
	case "42P01", "42703":
		return fmt.Errorf("%w: %w", ErrNotFound, de)
	case "40P01", "40001":
		return fmt.Errorf("%w: %w", ErrRetryable, de)
	default:
		return de
	}
}

var (
	ErrAlreadyExists = errors.New("already exists") // 23505
	ErrNotFound      = errors.New("not found")      // 42P01 (наш кейс “объект не найден”)
	ErrInvalid       = errors.New("invalid input")  // 23502, 23503, 23514, 22P02, 22023
	ErrRetryable     = errors.New("retryable")      // 40P01, 40001
)

type DomainError struct {
	Code, Message, Detail, Hint string
}

func (e *DomainError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Message, e.Detail, e.Code)
	}
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}
