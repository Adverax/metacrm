package validation

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	RuleTypeNotNil = "not_nil"
)

// ErrNotNilRequired is the error that returns when a value is Nil.
var ErrNotNilRequired = NewError("validation_not_nil_required", "is required")

// NotNil is a validation rule that checks if a value is not nil.
// NotNil only handles types including interface, pointer, slice, and map.
// All other types are considered valid.
var NotNil = notNilRule{condition: true}

type notNilRule struct {
	condition bool
	err       Error
}

func (r notNilRule) RuleType() RuleType {
	return RuleTypeNotNil
}

func (r notNilRule) MarshalJSON() ([]byte, error) {
	return []byte(`{}`), nil
}

func (r *notNilRule) UnmarshalJSON([]byte) error {
	return nil
}

// Validate checks if the given value is valid or not.
func (r notNilRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	_, isNil := Indirect(value)
	if isNil {
		if r.err != nil {
			return r.err
		}
		return ErrNotNilRequired
	}
	return nil
}

// Error sets the error message for the rule.
func (r notNilRule) Error(message string) notNilRule {
	if r.err == nil {
		r.err = ErrNotNilRequired
	}
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r notNilRule) ErrorObject(err Error) notNilRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r notNilRule) When(condition bool) notNilRule {
	r.condition = condition
	return r
}

func init() {
	RegisterUnmarshaller(RuleTypeNotNil, func(data []byte) (RuleEx, error) {
		rule := NotNil
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
