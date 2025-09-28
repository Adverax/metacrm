package validation

import (
	"context"
	"reflect"
)

// ErrUniqueListDuplicate is the error that returns in case of an duplication value for "unique" rule.
var ErrUniqueListDuplicate = NewError("validation_unique_list_duplication", "must have unique values")

// UniqueList returns a validation rule that checks if a value can be unique in the given list of values.
// reflect.DeepEqual() will be used to determine if two values are equal.
// For more details please refer to https://golang.org/pkg/reflect/#DeepEqual
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func UniqueList() UniqueListRule {
	return UniqueListRule{
		condition: true,
		err:       ErrUniqueListDuplicate,
	}
}

// UniqueListRule is a validation rule that validates if a value can be unique in the given list of values.
type UniqueListRule struct {
	condition bool
	err       Error
}

// Validate checks if the given value is valid or not.
func (r UniqueListRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	list := reflect.ValueOf(value)
	if list.Kind() != reflect.Slice && list.Kind() != reflect.Array {
		return nil
	}

	for i := 0; i < list.Len(); i++ {
		a := list.Index(i).Interface()
		for j := i + 1; j < list.Len(); j++ {
			b := list.Index(j).Interface()
			if reflect.DeepEqual(a, b) {
				return r.err
			}
		}
	}

	return nil
}

// Error sets the error message for the rule.
func (r UniqueListRule) Error(message string) UniqueListRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r UniqueListRule) ErrorObject(err Error) UniqueListRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r UniqueListRule) When(condition bool) UniqueListRule {
	r.condition = condition
	return r
}
