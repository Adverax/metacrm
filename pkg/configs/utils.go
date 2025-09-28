package configs

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/adverax/metacrm/pkg/types/convert"
	"reflect"
	"strings"
)

func Let(ctx context.Context, dst interface{}, src interface{}) error {
	handler := registry.Get(reflect.TypeOf(dst))
	if handler == nil {
		return nil
	}

	return handler.Let(ctx, dst, src)
}

// LetTyped assigns values from src to dst.
func LetTyped[T any](ctx context.Context, dst interface{}, src interface{}) error {
	if g, ok := src.(Getter[T]); ok {
		v, err := g.Get(ctx)
		if err != nil {
			return err
		}
		src = v
	} else {
		if d, ok := dst.(Importer); ok {
			return d.Import(ctx, src)
		}
	}

	if d, ok := dst.(Letter[T]); ok {
		val := reflect.ValueOf(src)
		tp := val.Type()
		if val.Type().ConvertibleTo(tp) {
			var v T
			value := val.Convert(tp)
			reflect.ValueOf(&v).Elem().Set(value)
			return d.Let(ctx, v)
		}
	}

	return nil
}

// Assign assigns values from src to dst.
func Assign(ctx context.Context, dst interface{}, src map[string]interface{}) error {
	dstValue := reflect.ValueOf(dst).Elem()
	dstType := dstValue.Type()

	for i := 0; i < dstValue.NumField(); i++ {
		field := dstValue.Field(i)
		fieldType := dstType.Field(i)

		if !fieldType.IsExported() {
			continue
		}
		if !field.CanSet() {
			continue
		}

		raw := fieldType.Tag.Get("config")
		if raw == "-" {
			continue
		}

		tags := ParseTags(raw)

		name := strings.ToLower(fieldType.Name)
		if tag, ok := tags["name"]; ok {
			name = tag
		}

		if value, ok := src[name]; ok {
			kind := field.Kind()
			switch kind {
			case reflect.Interface:
				err := Let(ctx, field.Interface(), value)
				if err != nil {
					return err
				}
			case reflect.Struct:
				if val, ok := value.(map[string]interface{}); ok {
					err := Assign(ctx, field.Addr().Interface(), val)
					if err != nil {
						return err
					}
				}
			default:
				if v, ok := convert.To(value, field.Type()); ok {
					field.Set(v)
				}
			}
		}
	}
	return nil
}

func override(a, b map[string]interface{}) {
	for k, v := range b {
		if av, ok := a[k]; ok {
			if reflect.TypeOf(v) == reflect.TypeOf(av) {
				switch v.(type) {
				case map[string]interface{}:
					override(av.(map[string]interface{}), v.(map[string]interface{}))
				case []interface{}:
					a[k] = v
				default:
					a[k] = v
				}
			} else {
				a[k] = v
			}
		} else {
			a[k] = v
		}
	}
}

func hashOf(data map[string]interface{}) string {
	bs, _ := json.MarshalIndent(data, "", "")
	return digestOf(bs)
}

func digestOf(bs []byte) string {
	return fmt.Sprintf("%x", md5.Sum(bs))
}

func ParseTags(tags string) map[string]string {
	list := strings.Split(tags, ",")
	res := make(map[string]string)
	for i, tag := range list {
		if tag == "" {
			continue
		}

		var frames []string
		frames = strings.Split(tag, "=")
		if i == 0 {
			if len(frames) == 1 {
				res["name"] = frames[0]
				continue
			}
		}

		if len(frames) == 1 {
			res[tag] = ""
		} else {
			res[frames[0]] = frames[1]
		}
	}

	return res
}
