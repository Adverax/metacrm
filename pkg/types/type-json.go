package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adverax/metacrm/pkg/types/convert"
)

type JsonType struct {
}

func (that *JsonType) Is(value interface{}) bool {
	switch value.(type) {
	case json.RawMessage:
	default:
		return false
	}

	return true
}

func (that *JsonType) IsAll(values []interface{}) bool {
	return IsAll(values, that)
}

func (that *JsonType) Get(ctx context.Context, getter Getter, name string, defVal json.RawMessage) (res json.RawMessage, err error) {
	val, err := getter.GetProperty(ctx, name)
	if err != nil {
		if errors.Is(err, ErrNoMatch) {
			return defVal, nil
		}
		return
	}
	if val == nil {
		return defVal, nil
	}
	if res, ok := that.TryCast(val); ok {
		return res, nil
	}
	return nil, fmt.Errorf("can not convert value %v into duration with key %q", val, name)
}

func (that *JsonType) TryCast(value interface{}) (json.RawMessage, bool) {
	return convert.ToJson(value)
}

func (that *JsonType) Cast(v interface{}, defaults json.RawMessage) json.RawMessage {
	if vv, ok := that.TryCast(v); ok {
		return vv
	}
	return defaults
}

func (that *JsonType) CastAll(values []interface{}) []json.RawMessage {
	result := make([]json.RawMessage, len(values))
	for i, value := range values {
		result[i] = that.Cast(value, nil)
	}
	return result
}
