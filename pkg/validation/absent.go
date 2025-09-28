package validation

import (
	"context"
)

var (
	// ErrNil is the error that returns when a value is not nil.
	ErrNil = NewError("validation_nil", "must be blank")
	// ErrEmpty is the error that returns when a not nil value is not empty.
	ErrEmpty = NewError("validation_empty", "must be blank")
)

const (
	RuleTypeNil   RuleType = "nil"
	RuleTypeEmpty RuleType = "empty"
)

// Nil is a validation rule that checks if a value is nil.
// It is the opposite of NotNil rule
var Nil = absentRule{
	absentRuleOptions: absentRuleOptions{
		SkipNil: false,
	},
	condition: true,
}

// Empty checks if a not nil value is empty.
var Empty = absentRule{
	absentRuleOptions: absentRuleOptions{
		SkipNil: true,
	},
	condition: true,
}

type absentRuleOptions struct {
	SkipNil bool `json:"skip_nil,omitempty"`
}

type absentRule struct {
	absentRuleOptions
	condition bool
	err       Error
}

func (r absentRule) RuleType() RuleType {
	if r.SkipNil {
		return RuleTypeEmpty
	}
	return RuleTypeNil
}

func (r absentRule) MarshalJSON() ([]byte, error) {
	return []byte(`{}`), nil
}

func (r *absentRule) UnmarshalJSON(data []byte) error {
	return nil
}

// Validate checks if the given value is valid or not.
func (r absentRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if !r.SkipNil && !isNil || r.SkipNil && !isNil && !IsEmpty(value) {
		if r.err != nil {
			return r.err
		}
		if r.SkipNil {
			return ErrEmpty
		}
		return ErrNil
	}

	return nil
}

// When sets the condition that determines if the validation should be performed.
func (r absentRule) When(condition bool) absentRule {
	r.condition = condition
	return r
}

// Error sets the error message for the rule.
func (r absentRule) Error(message string) absentRule {
	if r.err == nil {
		if r.SkipNil {
			r.err = ErrEmpty
		} else {
			r.err = ErrNil
		}
	}
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r absentRule) ErrorObject(err Error) absentRule {
	r.err = err
	return r
}

func init() {
	RegisterUnmarshaller(RuleTypeNil, func(data []byte) (RuleEx, error) {
		rule := Nil
		return &rule, nil
	})
	RegisterUnmarshaller(RuleTypeEmpty, func(data []byte) (RuleEx, error) {
		rule := Empty
		return &rule, nil
	})
}
