package validation

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	RuleTypeRequired      = "required"
	RuleTypeNirOrNotEmpty = "nil_or_not_empty"
)

var (
	// ErrRequired is the error that returns when a value is required.
	ErrRequired = NewError("validation_required", "cannot be blank")
	// ErrNilOrNotEmpty is the error that returns when a value is not nil and is empty.
	ErrNilOrNotEmpty = NewError("validation_nil_or_not_empty_required", "cannot be blank")
)

// Required is a validation rule that checks if a value is not empty.
// A value is considered not empty if
// - integer, float: not zero
// - bool: true
// - string, array, slice, map: len() > 0
// - interface, pointer: not nil and the referenced value is not empty
// - any other types
var Required = RequiredRule{
	requiredRuleOptions: requiredRuleOptions{
		SkipNil: false,
	},
	condition: true,
}

// NilOrNotEmpty checks if a value is a nil pointer or a value that is not empty.
// NilOrNotEmpty differs from Required in that it treats a nil pointer as valid.
var NilOrNotEmpty = RequiredRule{
	requiredRuleOptions: requiredRuleOptions{
		SkipNil: true,
	},
	condition: true,
}

type requiredRuleOptions struct {
	SkipNil bool `json:"skip_nil,omitempty"`
}

// RequiredRule is a rule that checks if a value is not empty.
type RequiredRule struct {
	requiredRuleOptions
	condition bool
	err       Error
}

func (r RequiredRule) RuleType() RuleType {
	if r.SkipNil {
		return RuleTypeNirOrNotEmpty
	}
	return RuleTypeRequired
}

func (r *RequiredRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.requiredRuleOptions)
}

func (r *RequiredRule) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.requiredRuleOptions); err != nil {
		return err
	}

	r.condition = true
	if r.SkipNil {
		r.err = ErrNilOrNotEmpty
	} else {
		r.err = ErrRequired
	}

	return nil
}

// Validate checks if the given value is valid or not.
func (r RequiredRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if r.SkipNil && !isNil && IsEmpty(value) || !r.SkipNil && (isNil || IsEmpty(value)) {
		if r.err != nil {
			return r.err
		}
		if r.SkipNil {
			return ErrNilOrNotEmpty
		}
		return ErrRequired
	}

	return nil
}

// When sets the condition that determines if the validation should be performed.
func (r RequiredRule) When(condition bool) RequiredRule {
	r.condition = condition
	return r
}

// Error sets the error message for the rule.
func (r RequiredRule) Error(message string) RequiredRule {
	if r.err == nil {
		if r.SkipNil {
			r.err = ErrNilOrNotEmpty
		} else {
			r.err = ErrRequired
		}
	}
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r RequiredRule) ErrorObject(err Error) RequiredRule {
	r.err = err
	return r
}

func init() {
	RegisterUnmarshaller(RuleTypeRequired, func(data []byte) (RuleEx, error) {
		rule := Required
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
	RegisterUnmarshaller(RuleTypeNirOrNotEmpty, func(data []byte) (RuleEx, error) {
		rule := NilOrNotEmpty
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
