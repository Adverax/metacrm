package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/adverax/metacrm/pkg/types/convert"
)

type BooleanType struct{}

func (that *BooleanType) Is(value interface{}) bool {
	switch value.(type) {
	case bool:
	default:
		return false
	}

	return true
}

func (that *BooleanType) IsAll(values []interface{}) bool {
	return IsAll(values, that)
}

func (that *BooleanType) Get(ctx context.Context, getter Getter, name string, defVal bool) (res bool, err error) {
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
	return false, fmt.Errorf("can not convert value %v into boolean with key %q", val, name)
}

func (that *BooleanType) TryCast(value interface{}) (bool, bool) {
	return convert.ToBoolean(value)
}

func (that *BooleanType) Cast(v interface{}, defaults bool) bool {
	if vv, ok := that.TryCast(v); ok {
		return vv
	}
	return defaults
}

func (that *BooleanType) CastAll(values []interface{}) []bool {
	result := make([]bool, len(values))
	for i, value := range values {
		result[i] = that.Cast(value, false)
	}
	return result
}
