package validation

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	RuleTypeCriteria RuleType = "criteria"
)

// ErrCriteriaIsNotMatch is the error that returns in case of an invalid value for "in" rule.
var ErrCriteriaIsNotMatch = NewError("validation_value_do_not_match_criteria", "value don't match criteria")

// Criteria returns a validation rule that checks if a value is match given value
func Criteria(expr string) CriteriaRule {
	return CriteriaRule{
		criteriaRuleOptions: criteriaRuleOptions{
			Expression: expr,
		},
		condition: true,
		errMatch:  ErrCriteriaIsNotMatch,
		errs:      defaultCELErrors,
	}
}

type criteriaRuleOptions struct {
	Expression string `json:"expr"`
}

// CriteriaRule is a validation rule that validates if a value match given value.
type CriteriaRule struct {
	criteriaRuleOptions
	condition bool
	errMatch  Error     // Error to return when the value does not match the criteria
	errs      celErrors // CEL errors for environment, compilation, and program creation
}

func (r CriteriaRule) RuleType() RuleType {
	return RuleTypeCriteria
}

func (r *CriteriaRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.criteriaRuleOptions)
}

func (r *CriteriaRule) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.criteriaRuleOptions); err != nil {
		return err
	}

	r.condition = true
	r.errMatch = ErrCriteriaIsNotMatch
	r.errs = defaultCELErrors

	return nil
}

// Validate checks if the given value is valid or not.
func (r CriteriaRule) Validate(ctx context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	checked, err := validateCriteria(ctx, r.Expression, value, &r.errs)
	if err != nil {
		return err
	}

	if checked {
		return nil
	}

	return r.errMatch
}

// Error sets the error message for the rule.
func (r CriteriaRule) Error(message string) CriteriaRule {
	r.errMatch = r.errMatch.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r CriteriaRule) ErrorObject(err Error) CriteriaRule {
	r.errMatch = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r CriteriaRule) When(condition bool) CriteriaRule {
	r.condition = condition
	return r
}

func init() {
	RegisterUnmarshaller(RuleTypeCriteria, func(data []byte) (RuleEx, error) {
		rule := Criteria("")
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
