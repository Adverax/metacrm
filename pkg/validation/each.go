package validation

import (
	"context"
	"errors"
	"reflect"
	"strconv"
)

// ErrInvalidKey is the error returned when a key in a map is not valid according to the specified rules.
var ErrInvalidKey = NewError("validation_invalid_key", "key is not valid")

// Each returns a validation rule that loops through an iterable (map, slice or array)
// and validates each value inside with the provided rules.
// An empty iterable is considered valid. Use the Required rule to make sure the iterable is not empty.
func Each(rules ...Rule) EachRule {
	return EachRule{
		condition:     true,
		valRules:      rules,
		errInvalidKey: ErrInvalidKey,
	}
}

// EachRule is a validation rule that validates elements in a map/slice/array using the specified list of rules.
type EachRule struct {
	condition     bool
	keyRules      []Rule
	valRules      []Rule
	errInvalidKey Error
}

// Validate loops through the given iterable and calls the Ozzo ValidateWithContext() method for each value.
func (r EachRule) Validate(ctx context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	var level ErrorLevel

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			err := Validate(ctx, key.Interface(), r.keyRules...)
			if err != nil {
				if !IsValidationError(err) {
					return err
				}
				k := r.getString(key)
				level.AddChildError(k, r.errInvalidKey.SetCause(err))
				continue
			}
			val := r.getInterface(v.MapIndex(key))
			err = Validate(ctx, val, r.valRules...)
			if err != nil {
				if !IsValidationError(err) {
					return err
				}
				level.AddChildError(r.getString(key), EnsureLevel(err))
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			val := r.getInterface(v.Index(i))
			err := Validate(ctx, val, r.valRules...)
			if err != nil {
				if !IsValidationError(err) {
					return err
				}
				level.AddChildError(strconv.Itoa(i), err)
			}
		}
	default:
		return errors.New("must be an iterable (map, slice or array)")
	}

	return level.Result()
}

func (r EachRule) getInterface(value reflect.Value) interface{} {
	return value.Interface()

	//switch value.Kind() {
	//case reflect.Ptr, reflect.Interface:
	//	if value.IsNil() {
	//		return nil
	//	}
	//	return value.Elem().Interface()
	//default:
	//	return value.Interface()
	//}
}

func (r EachRule) getString(value reflect.Value) string {
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		if value.IsNil() {
			return ""
		}
		return value.Elem().String()
	default:
		return value.String()
	}
}

// When sets the condition that determines if the validation should be performed.
func (r EachRule) When(condition bool) EachRule {
	r.condition = condition
	return r
}

// Key sets the rules to validate the keys of the map.
func (r EachRule) Key(rules ...Rule) EachRule {
	r.keyRules = rules
	return r
}
