package types

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/adverax/metacrm/pkg/types/convert"
)

type DurationType struct{}

func (that *DurationType) Is(value interface{}) bool {
	switch value.(type) {
	case time.Duration:
	case string:
	default:
		return false
	}

	return true
}

func (that *DurationType) IsAll(values []interface{}) bool {
	return IsAll(values, that)
}

func (that *DurationType) Get(ctx context.Context, getter Getter, name string, defVal time.Duration) (res time.Duration, err error) {
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
	return 0, fmt.Errorf("can not convert value %v into duration with key %q", val, name)
}

func (that *DurationType) TryCast(value interface{}) (time.Duration, bool) {
	return convert.ToDuration(value)
}

func (that *DurationType) Cast(v interface{}, defaults time.Duration) time.Duration {
	if vv, ok := that.TryCast(v); ok {
		return vv
	}
	return defaults
}

func (that *DurationType) CastAll(values []interface{}) []time.Duration {
	result := make([]time.Duration, len(values))
	for i, value := range values {
		result[i] = that.Cast(value, 0)
	}
	return result
}

func init() {
	convert.Register(
		reflect.TypeOf(time.Duration(0)),
		func(value interface{}) (reflect.Value, bool) {
			switch vv := value.(type) {
			case time.Duration:
				return reflect.ValueOf(vv), true
			case string:
				v, err := time.ParseDuration(vv)
				if err == nil {
					return reflect.ValueOf(v), true
				}
			}
			return reflect.Value{}, false
		},
	)
}
