package validation

import (
	"context"
	"encoding/json"
	"reflect"
)

var PrimitiveErrors = map[string]Error{
	new(Integer).GetName(): ErrValidationIntegerUnmarshal,
	new(Float).GetName():   ErrValidationFloatUnmarshal,
	new(String).GetName():  ErrValidationStringUnmarshal,
	new(Boolean).GetName(): ErrValidationBooleanUnmarshal,
}

type Integer struct {
	Primitive[int64]
}

type Float struct {
	Primitive[float64]
}

type String struct {
	Primitive[string]
}

type Boolean struct {
	Primitive[bool]
}

type Primitive[T any] struct {
	Value T
	Valid bool
	Error error
}

func (that *Primitive[T]) GetName() string {
	return reflect.TypeOf(that).Elem().Name()
}

func (that *Primitive[T]) GetValue() interface{} {
	if !that.Valid {
		return nil
	}
	return that.Value
}

func (that *Primitive[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		that.Valid = false
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		that.Valid = false
		baseError := PrimitiveErrors[that.GetName()]
		that.Error = baseError.
			SetParams(map[string]interface{}{
				"value": string(data),
			})
		return nil
	}

	that.Value = value
	that.Valid = true
	return nil
}

func (that *Primitive[T]) MarshalJSON() ([]byte, error) {
	if !that.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(that.Value)
}

func (that *Primitive[T]) Validate(ctx context.Context) error {
	return that.Error
}

var (
	ErrValidationIntegerUnmarshal = NewError("validation_integer_unmarshal", "failed to unmarshal integer value")
	ErrValidationFloatUnmarshal   = NewError("validation_float_unmarshal", "failed to unmarshal float value")
	ErrValidationStringUnmarshal  = NewError("validation_string_unmarshal", "failed to unmarshal string value")
	ErrValidationBooleanUnmarshal = NewError("validation_boolean_unmarshal", "failed to unmarshal boolean value")
)
