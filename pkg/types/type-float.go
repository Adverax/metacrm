package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adverax/metacrm/pkg/types/convert"
)

type FloatType struct{}

func (that *FloatType) Is(value interface{}) bool {
	switch value.(type) {
	case float32:
	case float64:
	case json.Number:
	default:
		return false
	}

	return true
}

func (that *FloatType) IsAll(values []interface{}) bool {
	return IsAll(values, that)
}

func (that *FloatType) Get(ctx context.Context, getter Getter, name string, defVal float64) (res float64, err error) {
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
	return 0, fmt.Errorf("can not convert value %v into float with key %q", val, name)
}

func (that *FloatType) TryCast(value interface{}) (float64, bool) {
	return convert.ToFloat64(value)
}

func (that *FloatType) Cast(v interface{}, defaults float64) float64 {
	if vv, ok := that.TryCast(v); ok {
		return vv
	}
	return defaults
}

func (that *FloatType) CastAll(values []interface{}) []float64 {
	result := make([]float64, len(values))
	for i, value := range values {
		result[i] = that.Cast(value, 0)
	}
	return result
}
