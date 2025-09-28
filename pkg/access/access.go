package access

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adverax/metacrm/pkg/types"
)

// Getter is abstract property props
type Getter interface {
	GetProperty(ctx context.Context, name string) (interface{}, error)
}

// Setter is abstract property setter
type Setter interface {
	SetProperty(ctx context.Context, name string, value interface{}) error
}

type GetterFunc func(ctx context.Context, name string) (interface{}, error)

func (fn GetterFunc) GetProperty(ctx context.Context, name string) (interface{}, error) {
	return fn(ctx, name)
}

type SetterFunc func(ctx context.Context, name string, value interface{}) error

func (fn SetterFunc) SetProperty(ctx context.Context, name string, value interface{}) error {
	return fn(ctx, name, value)
}

// GetterSetter is abstract property props & setter
type GetterSetter interface {
	Getter
	Setter
}

// Reader is abstract typed property props
type Reader interface {
	GetBoolean(ctx context.Context, name string, defVal bool) (bool, error)
	GetInteger(ctx context.Context, name string, defVal int64) (int64, error)
	GetFloat(ctx context.Context, name string, defVal float64) (float64, error)
	GetString(ctx context.Context, name string, defVal string) (string, error)
	GetDuration(ctx context.Context, name string, defVal time.Duration) (time.Duration, error)
	GetJson(ctx context.Context, name string, defVal json.RawMessage) (json.RawMessage, error)
}

// Writer is abstract typed property setter
type Writer interface {
	SetBoolean(ctx context.Context, name string, value bool) error
	SetInteger(ctx context.Context, name string, value int64) error
	SetFloat(ctx context.Context, name string, value float64) error
	SetString(ctx context.Context, name string, value string) error
	SetDuration(ctx context.Context, name string, value time.Duration) error
	SetJson(ctx context.Context, name string, value json.RawMessage) error
}

type ReaderGetter interface {
	Reader
	Getter
}

// ReaderWriter is abstract typed property reader & writer
type ReaderWriter interface {
	Getter
	Setter
	Reader
	Writer
}

type ReaderWriterEx interface {
	ReaderWriter
	Transaction(ctx context.Context, action func(ctx context.Context, rw ReaderWriter) error) error
}

func GetValue[T any](ctx context.Context, getter Getter, key string) (val T, err error) {
	v, err := getter.GetProperty(ctx, key)
	if err != nil {
		return val, err
	}

	if vv, ok := v.(T); ok {
		return vv, nil
	}

	return val, fmt.Errorf("can't typecast %v as %v", v, val)
}

func GetValueFromContext[T any](ctx context.Context, key string) (val T, err error) {
	return GetValue[T](ctx, ctxGetter, key)
}

func SetValue[T any](ctx context.Context, setter Setter, key string, val T) error {
	return setter.SetProperty(ctx, key, val)
}

var ctxGetter = &contextGetter{}

type contextGetter struct {
}

func (that *contextGetter) GetProperty(ctx context.Context, name string) (interface{}, error) {
	val := ctx.Value(name)
	if val == nil {
		return nil, types.ErrNoMatch
	}
	return val, nil
}
