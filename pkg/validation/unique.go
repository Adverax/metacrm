package validation

import (
	"context"
	"reflect"
)

// ErrUniqueDuplicate is the error that returns in case of an duplication value for "unique" rule.
var ErrUniqueDuplicate = NewError("validation_unique_duplication", "must be an unique value")

// Unique returns a validation rule that checks if a value can be unique in the given list of values.
// reflect.DeepEqual() will be used to determine if two values are equal.
// For more details please refer to https://golang.org/pkg/reflect/#DeepEqual
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func Unique(values ...interface{}) UniqueRule {
	return UniqueRule{
		condition: true,
		elements:  values,
		err:       ErrUniqueDuplicate,
	}
}

// UniqueRule is a validation rule that validates if a value can be unique in the given list of values.
type UniqueRule struct {
	condition bool
	elements  []interface{}
	err       Error
}

// Validate checks if the given value is valid or not.
func (r UniqueRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	var count int
	for _, e := range r.elements {
		if reflect.DeepEqual(e, value) {
			count++
			if count > 1 {
				return r.err
			}
		}
	}

	return nil
}

// Error sets the error message for the rule.
func (r UniqueRule) Error(message string) UniqueRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r UniqueRule) ErrorObject(err Error) UniqueRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r UniqueRule) When(condition bool) UniqueRule {
	r.condition = condition
	return r
}
