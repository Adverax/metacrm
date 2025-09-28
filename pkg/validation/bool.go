package validation

import "context"

// ErrTrueInvalid is the error that returns in case of changed value.
var ErrTrueInvalid = NewError("validation_true_invalid", "must be in a true value")

// ErrFalseInvalid is the error that returns in case of changed value.
var ErrFalseInvalid = NewError("validation_false_invalid", "must be in a false value")

// True returns a validation rule that checks if a value true.
// This rule should only be used for validating strings and byte slices, or a validation error will be reported.
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func True(actual bool) BoolRule {
	return BoolRule{
		condition: true,
		actual:    actual,
		expected:  true,
		err:       ErrTrueInvalid,
	}
}

// False returns a validation rule that checks if a value false.
// This rule should only be used for validating strings and byte slices, or a validation error will be reported.
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func False(actual any) BoolRule {
	return BoolRule{
		condition: true,
		actual:    actual,
		expected:  false,
		err:       ErrFalseInvalid,
	}
}

// BoolRule is a validation rule that checks if a value bool.
type BoolRule struct {
	condition bool
	actual    any
	expected  bool
	err       Error
}

// Validate checks if the given value is valid or not.
func (r BoolRule) Validate(context.Context, interface{}) error {
	if !r.condition {
		return nil
	}

	if r.actual == r.expected {
		return nil
	}

	return r.err
}

// Error sets the error message for the rule.
func (r BoolRule) Error(message string) BoolRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r BoolRule) ErrorObject(err Error) BoolRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r BoolRule) When(condition bool) BoolRule {
	r.condition = condition
	return r
}
