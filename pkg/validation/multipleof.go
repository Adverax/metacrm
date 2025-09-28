package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	RuleTypeMultipleOf = "multiple_of"
)

// ErrMultipleOfInvalid is the error that returns when a value is not multiple of a base.
var ErrMultipleOfInvalid = NewError("validation_multiple_of_invalid", "must be multiple of {{.base}}")

// MultipleOf returns a validation rule that checks if a value is a multiple of the "base" value.
// Note that "base" should be of integer type.
func MultipleOf(base interface{}) MultipleOfRule {
	return MultipleOfRule{
		multipleOfRuleOptions: multipleOfRuleOptions{
			Base: base,
		},
		condition: true,
		err:       ErrMultipleOfInvalid,
	}
}

type multipleOfRuleOptions struct {
	Base interface{} `json:"base"` // Base value to check against
}

// MultipleOfRule is a validation rule that checks if a value is a multiple of the "base" value.
type MultipleOfRule struct {
	multipleOfRuleOptions
	condition bool
	err       Error
}

func (r MultipleOfRule) RuleType() RuleType {
	return RuleTypeMultipleOf
}

func (r MultipleOfRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.multipleOfRuleOptions)
}

func (r *MultipleOfRule) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &r.multipleOfRuleOptions)
	if err != nil {
		return err
	}

	r.condition = true
	r.err = ErrMultipleOfInvalid

	return nil
}

// Error sets the error message for the rule.
func (r MultipleOfRule) Error(message string) MultipleOfRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r MultipleOfRule) ErrorObject(err Error) MultipleOfRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r MultipleOfRule) When(condition bool) MultipleOfRule {
	r.condition = condition
	return r
}

// Validate checks if the value is a multiple of the "base" value.
func (r MultipleOfRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	rv := reflect.ValueOf(r.Base)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := ToInt(value)
		if err != nil {
			return err
		}
		if v%rv.Int() == 0 {
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, err := ToUint(value)
		if err != nil {
			return err
		}

		if v%rv.Uint() == 0 {
			return nil
		}
	default:
		return fmt.Errorf("type not supported: %v", rv.Type())
	}

	return r.err.SetParams(map[string]interface{}{"base": r.Base})
}

func init() {
	RegisterUnmarshaller(RuleTypeMultipleOf, func(data []byte) (RuleEx, error) {
		rule := MultipleOf(nil)
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
