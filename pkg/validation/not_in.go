package validation

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	NotInRuleType = "not_in"
)

// ErrNotInInvalid is the error that returns when a value is in a list.
var ErrNotInInvalid = NewError("validation_not_in_invalid", "must not be in list")

type notInRuleOptions struct {
	Elements []interface{} `json:"elements"` // List of values to check against
}

// NotIn returns a validation rule that checks if a value is absent from the given list of values.
// Note that the value being checked and the possible range of values must be of the same type.
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func NotIn(values ...interface{}) NotInRule {
	return NotInRule{
		notInRuleOptions: notInRuleOptions{
			Elements: values,
		},
		condition: true,
		err:       ErrNotInInvalid,
	}
}

// NotInRule is a validation rule that checks if a value is absent from the given list of values.
type NotInRule struct {
	notInRuleOptions
	condition bool
	err       Error
}

func (r NotInRule) RuleType() RuleType {
	return NotInRuleType
}

func (r NotInRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.notInRuleOptions)
}

func (r *NotInRule) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.notInRuleOptions); err != nil {
		return err
	}

	r.condition = true
	r.err = ErrNotInInvalid

	return nil
}

// Validate checks if the given value is valid or not.
func (r NotInRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	for _, e := range r.Elements {
		if e == value {
			return r.err
		}
	}
	return nil
}

// Error sets the error message for the rule.
func (r NotInRule) Error(message string) NotInRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r NotInRule) ErrorObject(err Error) NotInRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r NotInRule) When(condition bool) NotInRule {
	r.condition = condition
	return r
}

func init() {
	RegisterUnmarshaller(NotInRuleType, func(data []byte) (RuleEx, error) {
		rule := NotIn()
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
