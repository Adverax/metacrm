package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
)

const (
	MatchRuleType = "match"
)

// ErrMatchInvalid is the error that returns in case of invalid format.
var ErrMatchInvalid = NewError("validation_match_invalid", "must be in a valid format")

// Match returns a validation rule that checks if a value matches the specified regular expression.
// This rule should only be used for validating strings and byte slices, or a validation error will be reported.
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func Match(re *regexp.Regexp) MatchRule {
	return MatchRule{
		condition: true,
		re:        re,
		err:       ErrMatchInvalid,
	}
}

type matchRuleOptions struct {
	ReSource string `json:"re,omitempty"`
}

// MatchRule is a validation rule that checks if a value matches the specified regular expression.
type MatchRule struct {
	matchRuleOptions
	condition bool
	re        *regexp.Regexp
	err       Error
}

func (r MatchRule) RuleType() RuleType {
	return MatchRuleType
}

func (r *MatchRule) MarshalJSON() ([]byte, error) {
	if r.re == nil {
		r.ReSource = ""
	} else {
		r.ReSource = r.re.String()
	}
	return json.Marshal(r.matchRuleOptions)
}

func (r *MatchRule) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.matchRuleOptions); err != nil {
		return err
	}

	r.condition = true
	if r.ReSource == "" {
		r.err = ErrMatchInvalid
		return nil
	}

	re, err := regexp.Compile(r.ReSource)
	if err != nil {
		r.err = ErrMatchInvalid.SetMessage(err.Error())
		return nil
	}
	r.re = re

	return nil
}

// Validate checks if the given value is valid or not.
func (r MatchRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil {
		return nil
	}

	isString, str, isBytes, bs := StringOrBytes(value)
	if isString && (str == "" || r.re.MatchString(str)) {
		return nil
	} else if isBytes && (len(bs) == 0 || r.re.Match(bs)) {
		return nil
	}
	return r.err
}

// Error sets the error message for the rule.
func (r MatchRule) Error(message string) MatchRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r MatchRule) ErrorObject(err Error) MatchRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r MatchRule) When(condition bool) MatchRule {
	r.condition = condition
	return r
}

func init() {
	RegisterUnmarshaller(MatchRuleType, func(data []byte) (RuleEx, error) {
		rule := Match(nil)
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
