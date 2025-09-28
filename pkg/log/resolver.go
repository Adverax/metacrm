package log

import (
	"context"
	"fmt"

	"github.com/adverax/metacrm/pkg/enums"
)

type contextLoggerType = int

const contextLoggerKey contextLoggerType = 0

type ContextMode int

func (that ContextMode) String() string {
	return ContextModes.DecodeOrDefault(that, "unknown")
}

const (
	ContextModeNone ContextMode = iota
	ContextModeTransparent
	ContextModeOpaque
)

var ContextModes = enums.New[ContextMode](
	map[ContextMode]string{
		ContextModeNone:        "",
		ContextModeTransparent: "transparent",
		ContextModeOpaque:      "opaque",
	},
)

type Resolver interface {
	Resolve(ctx context.Context) Logger
	NewContext(ctx context.Context) context.Context
}

type staticResolver struct {
	logger Logger
}

func (that *staticResolver) NewContext(ctx context.Context) context.Context {
	return ctx
}

func (that *staticResolver) Resolve(ctx context.Context) Logger {
	return that.logger
}

type opaqueResolver struct {
	logger Logger
}

func (that *opaqueResolver) NewContext(ctx context.Context) context.Context {
	return NewContext(ctx, that.logger)
}

func (that *opaqueResolver) Resolve(ctx context.Context) Logger {
	return GetLogger(ctx, that.logger)
}

type transparentResolver struct {
	logger Logger
}

func (that *transparentResolver) NewContext(ctx context.Context) context.Context {
	logger := GetLogger(ctx, nil)
	if logger == nil {
		return NewContext(ctx, that.logger)
	}

	return ctx
}

func (that *transparentResolver) Resolve(ctx context.Context) Logger {
	return GetLogger(ctx, that.logger)
}

func NewResolver(logger Logger, mode ContextMode) Resolver {
	switch mode {
	case ContextModeTransparent:
		return &transparentResolver{logger: logger}
	case ContextModeOpaque:
		return &opaqueResolver{logger: logger}
	default:
		return &staticResolver{logger: logger}
	}
}

// NewContext returns new context with logger
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextLoggerKey, logger)
}

func GetLogger(ctx context.Context, defVal Logger) Logger {
	val := ctx.Value(contextLoggerKey)
	if l, ok := val.(Logger); ok {
		return l
	}

	return defVal
}

// Resolve returns logger from context
var Resolve = func(ctx context.Context) Logger {
	log := GetLogger(ctx, nil)
	if log == nil {
		panic(fmt.Errorf("logger not found in context: %v", ctx))
	}

	return log
}
