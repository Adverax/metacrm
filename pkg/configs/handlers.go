package configs

import (
	"context"
	"reflect"
	"time"
)

type TypeHandler interface {
	Let(ctx context.Context, dst interface{}, src interface{}) error
}

type BooleanTypeHandler struct {
}

func (that *BooleanTypeHandler) Let(ctx context.Context, dst interface{}, src interface{}) error {
	return LetTyped[bool](ctx, dst, src)
}

type IntegerTypeHandler struct {
}

func (that *IntegerTypeHandler) Let(ctx context.Context, dst interface{}, src interface{}) error {
	return LetTyped[int64](ctx, dst, src)
}

type FloatTypeHandler struct {
}

func (that *FloatTypeHandler) Let(ctx context.Context, dst interface{}, src interface{}) error {
	return LetTyped[float64](ctx, dst, src)
}

type StringTypeHandler struct {
}

func (that *StringTypeHandler) Let(ctx context.Context, dst interface{}, src interface{}) error {
	return LetTyped[string](ctx, dst, src)
}

type DurationTypeHandler struct {
}

func (that *DurationTypeHandler) Let(ctx context.Context, dst interface{}, src interface{}) error {
	return LetTyped[time.Duration](ctx, dst, src)
}

type StringsTypeHandler struct {
}

func (that *StringsTypeHandler) Let(ctx context.Context, dst interface{}, src interface{}) error {
	return LetTyped[[]string](ctx, dst, src)
}

type TimeTypeHandler struct {
}

func (that *TimeTypeHandler) Let(ctx context.Context, dst interface{}, src interface{}) error {
	return LetTyped[time.Time](ctx, dst, src)
}

type Registry struct {
	handlers map[reflect.Type]TypeHandler
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[reflect.Type]TypeHandler),
	}
}

func (that *Registry) Register(tp reflect.Type, handler TypeHandler) *Registry {
	that.handlers[tp] = handler
	return that
}

func (that *Registry) Get(tp reflect.Type) TypeHandler {
	for t, h := range that.handlers {
		if tp.Implements(t) {
			return h
		}
	}
	return nil
}

func HandlerOf(tp reflect.Type) TypeHandler {
	return registry.Get(tp)
}

func RegisterHandler(tp reflect.Type, handler TypeHandler) {
	registry.Register(tp, handler)
}

var registry = NewRegistry().
	Register(reflect.TypeOf((*Boolean)(nil)).Elem(), &BooleanTypeHandler{}).
	Register(reflect.TypeOf((*Integer)(nil)).Elem(), &IntegerTypeHandler{}).
	Register(reflect.TypeOf((*Float)(nil)).Elem(), &FloatTypeHandler{}).
	Register(reflect.TypeOf((*String)(nil)).Elem(), &StringTypeHandler{}).
	Register(reflect.TypeOf((*Duration)(nil)).Elem(), &DurationTypeHandler{}).
	Register(reflect.TypeOf((*Strings)(nil)).Elem(), &StringsTypeHandler{}).
	Register(reflect.TypeOf((*Time)(nil)).Elem(), &TimeTypeHandler{})
