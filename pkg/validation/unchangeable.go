package validation

import (
	"context"
	"reflect"
)

// ErrUnchangeableInvalid is the error that returns in case of changed value.
var ErrUnchangeableInvalid = NewError("validation_unchangeable_invalid", "must be in a unchangeable value")

// Unchangeable returns a validation rule that checks if a value unchangeable.
// This rule should only be used for validating strings and byte slices, or a validation error will be reported.
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func Unchangeable(original any) UnchangeableRule {
	return UnchangeableRule{
		condition: true,
		original:  original,
		err:       ErrUnchangeableInvalid,
	}
}

// UnchangeableRule is a validation rule that checks if a value unchangeable.
type UnchangeableRule struct {
	condition bool
	original  any
	err       Error
}

// Validate checks if the given value is valid or not.
func (r UnchangeableRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	if IsChanged(r.original, value) {
		return r.err
	}

	return nil
}

// When sets the condition that determines if the validation should be performed.
func (r UnchangeableRule) When(condition bool) UnchangeableRule {
	r.condition = condition
	return r
}

// Error sets the error message for the rule.
func (r UnchangeableRule) Error(message string) UnchangeableRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r UnchangeableRule) ErrorObject(err Error) UnchangeableRule {
	r.err = err
	return r
}

// IsChanged checks if the original value is changed compared to the new value.
func IsChanged(original, value any) bool {
	if original == nil {
		return false
	}

	if reflect.DeepEqual(original, value) {
		return false
	}

	return true
}
