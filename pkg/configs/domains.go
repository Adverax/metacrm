package configs

import (
	"context"
	"sync"
	"time"
)

type Config interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type BaseConfig struct {
	sync.RWMutex
}

type Fetcher interface {
	Fetch() ([]byte, error)
}

type Source interface {
	Fetch() (map[string]interface{}, error)
}

type Converter interface {
	Convert(dst interface{}, src map[string]interface{}) error
}

type Boolean interface {
	Get(ctx context.Context) (bool, error)
}

type Integer interface {
	Get(ctx context.Context) (int64, error)
}

type Float interface {
	Get(ctx context.Context) (float64, error)
}

type String interface {
	Get(ctx context.Context) (string, error)
}

type Duration interface {
	Get(ctx context.Context) (time.Duration, error)
}

type Strings interface {
	Get(ctx context.Context) ([]string, error)
}

type Time interface {
	Get(ctx context.Context) (time.Time, error)
}

type Getter[T any] interface {
	Get(ctx context.Context) (T, error)
}

type Importer interface {
	Import(ctx context.Context, value interface{}) error
}

type Letter[T any] interface {
	Let(ctx context.Context, value T) error
}
