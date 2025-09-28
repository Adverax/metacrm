package types

import (
	"context"
	"errors"
)

// Getter is abstract property props
type Getter interface {
	GetProperty(ctx context.Context, name string) (interface{}, error)
}

// Setter is abstract property setter
type Setter interface {
	SetProperty(ctx context.Context, name string, value interface{}) error
}

type TypeChecker interface {
	Is(value interface{}) bool
}

type Typer[T any] interface {
	TypeChecker
	Get(ctx context.Context, getter Getter, name string, defVal T) (res T, err error)
	IsAll(values []interface{}) bool
	TryCast(value interface{}) (T, bool)
	Cast(value interface{}, defVal T) T
	CastAll(value []interface{}) []T
}

var (
	Boolean  = &BooleanType{}
	Integer  = &IntegerType{}
	Float    = &FloatType{}
	String   = &StringType{}
	Duration = &DurationType{}
	Json     = &JsonType{}
)

var (
	ErrNoMatch = errors.New("no match")
)

func IsAll(values []interface{}, typ TypeChecker) bool {
	for _, value := range values {
		if !typ.Is(value) {
			return false
		}
	}

	return true
}
