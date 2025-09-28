package access

import (
	"context"
	"encoding/json"
	"time"

	"github.com/adverax/metacrm/pkg/types"
)

type reader struct {
	Getter
}

func (that *reader) GetBoolean(
	ctx context.Context,
	name string,
	defVal bool,
) (res bool, err error) {
	return types.Boolean.Get(ctx, that, name, defVal)
}

func (that *reader) GetString(
	ctx context.Context,
	name string,
	defVal string,
) (res string, err error) {
	return types.String.Get(ctx, that, name, defVal)
}

func (that *reader) GetInteger(
	ctx context.Context,
	name string,
	defVal int64,
) (res int64, err error) {
	return types.Integer.Get(ctx, that, name, defVal)
}

func (that *reader) GetFloat(
	ctx context.Context,
	name string,
	defVal float64,
) (res float64, err error) {
	return types.Float.Get(ctx, that, name, defVal)
}

func (that *reader) GetDuration(
	ctx context.Context,
	name string,
	defVal time.Duration,
) (res time.Duration, err error) {
	return types.Duration.Get(ctx, that, name, defVal)
}

func (that *reader) GetJson(
	ctx context.Context,
	name string,
	defVal json.RawMessage,
) (res json.RawMessage, err error) {
	return types.Json.Get(ctx, that, name, defVal)
}

// NewReader is constructor for build Reader, based on the props
func NewReader(getter Getter) ReaderGetter {
	return &reader{
		Getter: getter,
	}
}
