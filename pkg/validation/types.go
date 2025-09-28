package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	RuleTypeType RuleType = "type"
)

// ErrTypeInvalid is the error that returns in case of an invalid type for "type" rule.
var ErrTypeInvalid = NewError("validation_type_invalid", "must be a valid type")

// Type returns a validation rule that checks if a value is match required type.
// reflect.DeepEqual() will be used to determine if two values are equal.
// For more details please refer to https://golang.org/pkg/reflect/#DeepEqual
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func Type(types ...string) TypeRule {
	return TypeRule{
		typeRuleOptions: typeRuleOptions{
			Types: types,
		},
		condition: true,
		err:       ErrTypeInvalid,
	}
}

type typeRuleOptions struct {
	Types []string `json:"types"`
}

// TypeRule is a validation rule that validates if a value has one of given list names of types.
type TypeRule struct {
	typeRuleOptions
	condition bool
	err       Error
}

func (r TypeRule) RuleType() RuleType {
	return RuleTypeType
}

func (r *TypeRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.typeRuleOptions)
}

func (r *TypeRule) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.typeRuleOptions); err != nil {
		return err
	}

	r.condition = true
	r.err = ErrInInvalid

	return nil
}

// Validate checks if the given value is valid or not.
func (r TypeRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		if isNil {
			return r.err
		}
		return nil
	}

	tp := reflect.TypeOf(value).String()

	for _, t := range r.Types {
		if t == tp {
			return nil
		}
	}

	return r.err
}

// Error sets the error message for the rule.
func (r TypeRule) Error(message string) TypeRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r TypeRule) ErrorObject(err Error) TypeRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r TypeRule) When(condition bool) TypeRule {
	r.condition = condition
	return r
}

func init() {
	RegisterUnmarshaller(RuleTypeType, func(data []byte) (RuleEx, error) {
		rule := Type()
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
