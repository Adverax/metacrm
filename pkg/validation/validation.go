// Package validation provides configurable and extensible rules for validating data of various types.
package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type (
	// Validatable is the interface indicating the type implementing it supports data validation.
	Validatable interface {
		// Validate validates the data and returns an error if validation fails.
		Validate(ctx context.Context) error
	}

	// Valuable is the interface indicating the type implementing it can return its value.
	Valuable interface {
		// GetValue returns the value of the object.
		GetValue() interface{}
	}

	// Rule represents a validation rule.
	Rule interface {
		// Validate validates a value and returns a value if validation fails.
		Validate(ctx context.Context, value interface{}) error
	}

	// RuleEx is an extended Rule interface that includes a Type method to return the type of the rule.
	// You must use it for serialization purposes.
	RuleEx interface {
		RuleType() RuleType
		json.Marshaler
		json.Unmarshaler
		Rule
	}

	// RuleFunc represents a validator function.
	// You may wrap it as a Rule by calling By().
	RuleFunc func(ctx context.Context, value interface{}) error
)

var (
	// ErrorTag is the struct tag name used to customize the error field name for a struct field.
	ErrorTag = "json"

	// Skip is a special validation rule that indicates all rules following it should be skipped.
	Skip = skipRule{skip: true}

	validatableType = reflect.TypeOf((*Validatable)(nil)).Elem()
)

// Validate validates the given value with the given context and returns the validation error, if any.
//
// Validate performs validation using the following steps:
//  1. For each rule, call its `ValidateWithContext()` to validate the value if the rule implements `RuleWithContext`.
//     Otherwise call `Validate()` of the rule. Return if any error is found.
//  2. If the value being validated implements `ValidatableWithContext`, call the value's `ValidateWithContext()`
//     and return with the validation result.
//  3. If the value being validated implements `Validatable`, call the value's `Validate()`
//     and return with the validation result.
//  4. If the value being validated is a map/slice/array, and the element type implements `Validatable`,
//     for each element call the element value's `Validate()`. Return with the validation result.
func Validate(ctx context.Context, value interface{}, rules ...Rule) error {
	var level ErrorLevel

	val := getInterface(reflect.ValueOf(value))
	for _, rule := range rules {
		if s, ok := rule.(skipRule); ok && s.skip {
			break
		}
		if err := rule.Validate(ctx, val); err != nil && !level.AddError(err) {
			return err
		}
	}

	if err := valid(ctx, value); err != nil && !level.AddError(err) {
		return err
	}

	return level.Result()
}

func validateRules(ctx context.Context, value interface{}, rules ...Rule) error {
	level := NewErrorLevel()

	for _, rule := range rules {
		if s, ok := rule.(skipRule); ok && s.skip {
			break
		}
		if err := rule.Validate(ctx, value); err != nil && !level.AddError(err) {
			return err
		}
	}

	if level.IsEmpty() {
		return nil
	}

	return level
}

func valid(ctx context.Context, value interface{}) error {
	rv := reflect.ValueOf(value)
	if (rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface) && rv.IsNil() {
		return nil
	}

	if v, ok := value.(Validatable); ok {
		p := ptrOf(value)
		if !IsVisited(ctx, p) {
			return v.Validate(withVisitContext(ctx, p))
		}
	}

	switch rv.Kind() {
	case reflect.Map:
		if rv.Type().Elem().Implements(validatableType) {
			return validateMap(ctx, rv)
		}
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Implements(validatableType) {
			return validateSlice(ctx, rv)
		}
	case reflect.Ptr, reflect.Interface:
		return Validate(ctx, rv.Elem().Interface())
	}

	return nil
}

// validateMap validates a map of validatable elements with the given context.
func validateMap(ctx context.Context, rv reflect.Value) error {
	errs := Errors{}
	for _, key := range rv.MapKeys() {
		if mv := rv.MapIndex(key).Interface(); mv != nil {
			if err := mv.(Validatable).Validate(ctx); err != nil {
				if !IsValidationError(err) {
					return err
				}
				errs[fmt.Sprintf("%v", key.Interface())] = EnsureLevel(err)
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// validateSlice validates a slice/array of validatable elements
func validateSlice(ctx context.Context, rv reflect.Value) error {
	errs := Errors{}
	l := rv.Len()
	for i := 0; i < l; i++ {
		if ev := rv.Index(i).Interface(); ev != nil {
			if err := ev.(Validatable).Validate(ctx); err != nil {
				if !IsValidationError(err) {
					return err
				}
				errs[strconv.Itoa(i)] = EnsureLevel(err)
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func getInterface(value reflect.Value) interface{} {
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Ptr:
		if value.IsNil() {
			return nil
		}
		if v, ok := value.Interface().(Valuable); ok {
			return v.GetValue()
		}
		return value.Elem().Interface()
	default:
		return value.Interface()
	}
}

type skipRule struct {
	skip bool
}

func (r skipRule) Validate(context.Context, interface{}) error {
	return nil
}

// When determines if all rules following it should be skipped.
func (r skipRule) When(condition bool) skipRule {
	r.skip = condition
	return r
}

type inlineRule struct {
	f RuleFunc
}

func (r *inlineRule) Validate(ctx context.Context, value interface{}) error {
	return r.f(ctx, value)
}

// By wraps a RuleFunc into a Rule.
func By(f RuleFunc) Rule {
	return &inlineRule{f: f}
}

func IsValidationError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(ErrorObject); ok {
		return true
	}

	if _, ok := err.(*ErrorLevel); ok {
		return true
	}

	if _, ok := err.(ErrorList); ok {
		return true
	}

	if _, ok := err.(Errors); ok {
		return true
	}

	return false
}

func ptrOf(value interface{}) uintptr {
	ref, isNil := Indirect(value)
	if isNil {
		return 0
	}

	vf := reflect.ValueOf(ref)
	switch vf.Kind() {
	case reflect.Map, reflect.Slice:
		return vf.Pointer()
	default:
		return 0
	}
}

type visitContextKey uintptr

func WithVisitContext(ctx context.Context, value interface{}) context.Context {
	ptr := ptrOf(value)
	if IsVisited(ctx, ptr) {
		return ctx
	}
	return withVisitContext(ctx, ptr)
}

func withVisitContext(ctx context.Context, ptr uintptr) context.Context {
	if ptr == 0 {
		return ctx
	}
	return context.WithValue(ctx, visitContextKey(ptr), struct{}{})
}

func IsVisited(ctx context.Context, ptr uintptr) bool {
	if ptr == 0 {
		return false
	}
	_, ok := ctx.Value(visitContextKey(ptr)).(struct{})
	return ok
}
