package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	RuleTypeIn RuleType = "in"
)

// ErrInInvalid is the error that returns in case of an invalid value for "in" rule.
var ErrInInvalid = NewError("validation_in_invalid", "must be a valid value")

// In returns a validation rule that checks if a value can be found in the given list of values.
// reflect.DeepEqual() will be used to determine if two values are equal.
// For more details please refer to https://golang.org/pkg/reflect/#DeepEqual
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func In(values ...interface{}) InRule {
	return InRule{
		inRuleOptions: inRuleOptions{
			Elements: values,
		},
		condition: true,
		err:       ErrInInvalid,
	}
}

type inRuleOptions struct {
	Elements []interface{} `json:"elements"`
}

// InRule is a validation rule that validates if a value can be found in the given list of values.
type InRule struct {
	inRuleOptions
	condition bool
	err       Error
}

func (r InRule) RuleType() RuleType {
	return RuleTypeIn
}

func (r *InRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.inRuleOptions)
}

func (r *InRule) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.inRuleOptions); err != nil {
		return err
	}

	r.condition = true
	r.err = ErrInInvalid

	return nil
}

// Validate checks if the given value is valid or not.
func (r InRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	for _, e := range r.Elements {
		if reflect.DeepEqual(e, value) {
			return nil
		}
	}

	return r.err
}

// Error sets the error message for the rule.
func (r InRule) Error(message string) InRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r InRule) ErrorObject(err Error) InRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r InRule) When(condition bool) InRule {
	r.condition = condition
	return r
}

func init() {
	RegisterUnmarshaller(RuleTypeIn, func(data []byte) (RuleEx, error) {
		rule := In()
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
