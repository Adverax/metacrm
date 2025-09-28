package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adverax/metacrm/pkg/types/convert"
)

type StringType struct{}

func (that *StringType) Is(value interface{}) bool {
	switch value.(type) {
	case string:
	case json.Number:
	default:
		return false
	}

	return true
}

func (that *StringType) IsAll(values []interface{}) bool {
	return IsAll(values, that)
}

func (that *StringType) Get(ctx context.Context, getter Getter, name string, defVal string) (res string, err error) {
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
	return "", fmt.Errorf("can not convert value %v into string with key %q", val, name)
}

func (that *StringType) TryCast(value interface{}) (string, bool) {
	return convert.ToString(value)
}

func (that *StringType) Cast(v interface{}, defaults string) string {
	if vv, ok := that.TryCast(v); ok {
		return vv
	}
	return defaults
}

func (that *StringType) CastAll(values []interface{}) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = that.Cast(value, "")
	}
	return result
}
