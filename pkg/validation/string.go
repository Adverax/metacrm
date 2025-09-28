package validation

import "context"

type stringValidator func(string) bool

// StringRule is a rule that checks a string variable using a specified stringValidator.
type StringRule struct {
	ruleType  RuleType
	condition bool
	validate  stringValidator
	err       Error
}

// NewStringRule creates a new validation rule using a function that takes a string value and returns a bool.
// The rule returned will use the function to check if a given string or byte slice is valid or not.
// An empty value is considered to be valid. Please use the Required rule to make sure a value is not empty.
func NewStringRule(t RuleType, validator stringValidator, message string) StringRule {
	return StringRule{
		validate:  validator,
		condition: true,
		err:       NewError("", message),
	}
}

// NewStringRuleWithError creates a new validation rule using a function that takes a string value and returns a bool.
// The rule returned will use the function to check if a given string or byte slice is valid or not.
// An empty value is considered to be valid. Please use the Required rule to make sure a value is not empty.
func NewStringRuleWithError(t RuleType, validator stringValidator, err Error) StringRule {
	return StringRule{
		validate:  validator,
		condition: true,
		err:       err,
	}
}

func (r StringRule) RuleType() RuleType {
	return r.ruleType
}

func (r StringRule) MarshalJSON() ([]byte, error) {
	return []byte(`{}`), nil
}

func (r StringRule) UnmarshalJSON(data []byte) error {
	return nil
}

// Error sets the error message for the rule.
func (r StringRule) Error(message string) StringRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r StringRule) ErrorObject(err Error) StringRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r StringRule) When(condition bool) StringRule {
	r.condition = condition
	return r
}

// Validate checks if the given value is valid or not.
func (r StringRule) Validate(_ context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	str, err := EnsureString(value)
	if err != nil {
		return err
	}

	if r.validate(str) {
		return nil
	}

	return r.err
}
