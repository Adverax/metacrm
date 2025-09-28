package access

import (
	"context"
	"encoding/json"
	"time"
)

type writer struct {
	Setter
}

func (that *writer) SetBoolean(
	ctx context.Context,
	name string,
	value bool,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that *writer) SetString(
	ctx context.Context,
	name string,
	value string,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that *writer) SetInteger(
	ctx context.Context,
	name string,
	value int64,
) error {
	//log.Println("SET INTEGER PROP", name, value)
	return that.SetProperty(ctx, name, value)
}

func (that *writer) SetFloat(
	ctx context.Context,
	name string,
	value float64,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that *writer) SetDuration(
	ctx context.Context,
	name string,
	value time.Duration,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that *writer) SetJson(
	ctx context.Context,
	name string,
	value json.RawMessage,
) error {
	return that.SetProperty(ctx, name, string(value))
}

// NewWriter is constructor for build Writer, based on the setter
func NewWriter(setter Setter) Writer {
	return &writer{
		Setter: setter,
	}
}
