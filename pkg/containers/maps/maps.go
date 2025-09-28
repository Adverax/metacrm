package maps

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/adverax/metacrm/pkg/types"
)

// Map is a standard map, that implements GetterSetter
type Map map[string]interface{}

func (that Map) Contains(name string) bool {
	_, err := that.GetProperty(context.Background(), name)
	return err == nil
}

func (that Map) String() string {
	return fmt.Sprintf("%#v", that)
}

func (that Map) GetProperty(
	_ context.Context,
	name string,
) (interface{}, error) {
	if strings.HasPrefix(name, "$.") {
		var path = strings.Split(name[2:], ".")

		if len(path) == 1 {
			if v, ok := that[name]; ok {
				return v, nil
			}
			return nil, types.ErrNoMatch
		}

		var val interface{} = that
		for _, key := range path {
			if vv, ok := val.(Map); ok {
				if v, ok := vv[key]; ok {
					val = v
				} else {
					return nil, types.ErrNoMatch
				}
			} else {
				return nil, types.ErrNoMatch
			}
		}

		return val, nil
	}

	if v, ok := that[name]; ok {
		return v, nil
	}

	return nil, types.ErrNoMatch
}

func (that Map) SetProperty(
	_ context.Context,
	name string,
	value interface{},
) error {
	if strings.HasPrefix(name, "$.") {
		path := strings.Split(name[2:], ".")
		if len(path) == 1 {
			that[name] = value
			return nil
		}

		var val interface{} = that
		for i := 0; i < len(path)-1; i++ {
			key := path[i]
			if vv, ok := val.(Map); ok {
				if v, ok := vv[key]; ok {
					val = v
				} else {
					vv[key] = make(Map)
					val = vv[key]
				}
			} else {
				return types.ErrNoMatch
			}
		}

		if vv, ok := val.(Map); ok {
			vv[path[len(path)-1]] = value
			return nil
		}

		return types.ErrNoMatch
	}

	that[name] = value
	return nil
}

func (that Map) ToBoolean(
	ctx context.Context,
	name string,
	defVal bool,
) bool {
	val, err := that.GetProperty(ctx, name)
	if err != nil || val == nil {
		return defVal
	}
	return types.Boolean.Cast(val, defVal)
}

func (that Map) ToString(
	ctx context.Context,
	name string,
	defVal string,
) string {
	val, err := that.GetProperty(ctx, name)
	if err != nil || val == nil {
		return defVal
	}
	return types.String.Cast(val, defVal)
}

func (that Map) ToInteger(
	ctx context.Context,
	name string,
	defVal int64,
) int64 {
	val, err := that.GetProperty(ctx, name)
	if err != nil || val == nil {
		return defVal
	}
	return types.Integer.Cast(val, defVal)
}

func (that Map) ToFloat(
	ctx context.Context,
	name string,
	defVal float64,
) float64 {
	val, err := that.GetProperty(ctx, name)
	if err != nil || val == nil {
		return defVal
	}
	return types.Float.Cast(val, defVal)
}

func (that Map) ToDuration(
	ctx context.Context,
	name string,
	defVal time.Duration,
) time.Duration {
	val, err := that.GetProperty(ctx, name)
	if err != nil || val == nil {
		return defVal
	}
	return types.Duration.Cast(val, defVal)
}

func (that Map) ToJson(
	ctx context.Context,
	name string,
	defVal json.RawMessage,
) json.RawMessage {
	val, err := that.GetProperty(ctx, name)
	if err != nil || val == nil {
		return defVal
	}
	return types.Json.Cast(val, defVal)
}

func (that Map) ToMap(
	_ context.Context,
	name string,
) Map {
	if mm, ok := that[name]; ok {
		if mmm, ok := mm.(Map); ok {
			return mmm
		}
		if mmm, ok := mm.(map[string]interface{}); ok {
			return mmm
		}
	}
	return nil
}

func (that Map) ToMaps(
	_ context.Context,
	name string,
) []Map {
	if mm, ok := that[name]; ok {
		if mmm, ok := mm.([]Map); ok {
			return mmm
		}
		if mmm, ok := mm.([]interface{}); ok {
			m1 := make([]Map, 0, len(mmm))
			for _, m2 := range mmm {
				if m3, ok := m2.(map[string]interface{}); ok {
					m1 = append(m1, m3)
				}
			}
			return m1
		}
	}
	return nil
}

func (that Map) ToSlice(
	_ context.Context,
	name string,
) []interface{} {
	if mm, ok := that[name]; ok {
		switch v := mm.(type) {
		case []interface{}:
			return v
		default:
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Slice {
				list := make([]interface{}, val.Len())
				for i := 0; i < val.Len(); i++ {
					list[i] = val.Index(i).Interface()
				}
				return list
			}
		}
	}
	return nil
}

func (that Map) GetBoolean(
	ctx context.Context,
	name string,
	defVal bool,
) (res bool, err error) {
	return types.Boolean.Get(ctx, that, name, defVal)
}

func (that Map) GetString(
	ctx context.Context,
	name string,
	defVal string,
) (res string, err error) {
	return types.String.Get(ctx, that, name, defVal)
}

func (that Map) GetInteger(
	ctx context.Context,
	name string,
	defVal int64,
) (res int64, err error) {
	return types.Integer.Get(ctx, that, name, defVal)
}

func (that Map) GetFloat(
	ctx context.Context,
	name string,
	defVal float64,
) (res float64, err error) {
	return types.Float.Get(ctx, that, name, defVal)
}

func (that Map) GetDuration(
	ctx context.Context,
	name string,
	defVal time.Duration,
) (res time.Duration, err error) {
	return types.Duration.Get(ctx, that, name, defVal)
}

func (that Map) GetJson(
	ctx context.Context,
	name string,
	defVal json.RawMessage,
) (res json.RawMessage, err error) {
	return types.Json.Get(ctx, that, name, defVal)
}

func (that Map) SetBoolean(
	ctx context.Context,
	name string,
	value bool,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that Map) SetString(
	ctx context.Context,
	name string,
	value string,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that Map) SetInteger(
	ctx context.Context,
	name string,
	value int64,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that Map) SetFloat(
	ctx context.Context,
	name string,
	value float64,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that Map) SetDuration(
	ctx context.Context,
	name string,
	value time.Duration,
) error {
	return that.SetProperty(ctx, name, value)
}

func (that Map) SetJson(
	ctx context.Context,
	name string,
	value json.RawMessage,
) error {
	return that.SetProperty(ctx, name, value)
}

// Scope is routine, that allow access to the branch of base Map as sub Map.
func (that Map) Scope(name string) Map {
	if mm, ok := that[name]; ok {
		if mmm, ok := mm.(Map); ok {
			return mmm
		}
	}
	return nil
}

// NewScope is routine, that allow access to the branch of base Map as sub Map.
func (that Map) NewScope(name string) Map {
	if mm, ok := that[name]; ok {
		if mmm, ok := mm.(Map); ok {
			return mmm
		}
	}
	mmm := make(Map)
	that[name] = mmm
	return mmm
}

func (that Map) Clone() Map {
	cp := make(Map)
	for k, v := range that {
		vm, ok := v.(Map)
		if ok {
			cp[k] = vm.Clone()
		} else {
			cp[k] = v
		}
	}

	return cp
}

func (that Map) IsEmpty() bool {
	return len(that) == 0
}
