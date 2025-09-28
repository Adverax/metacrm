package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adverax/metacrm/pkg/types/convert"
)

type IntegerType struct{}

func (that *IntegerType) Is(value interface{}) bool {
	switch value.(type) {
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
	case json.Number:
	default:
		return false
	}

	return true
}

func (that *IntegerType) IsAll(values []interface{}) bool {
	return IsAll(values, that)
}

func (that *IntegerType) Get(ctx context.Context, getter Getter, name string, defVal int64) (res int64, err error) {
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
	return 0, fmt.Errorf("can not convert value %v into integer with key %q", val, name)
}

func (that *IntegerType) TryCast(value interface{}) (int64, bool) {
	return convert.ToInt64(value)
}

func (that *IntegerType) Cast(v interface{}, defaults int64) int64 {
	if vv, ok := that.TryCast(v); ok {
		return vv
	}
	return defaults
}

func (that *IntegerType) CastAll(values []interface{}) []int64 {
	result := make([]int64, len(values))
	for i, value := range values {
		result[i] = that.Cast(value, 0)
	}
	return result
}
