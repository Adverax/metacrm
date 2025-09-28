package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	RuleTypeDate = "date"
)

var (
	// ErrDateInvalid is the error that returns in case of an invalid date.
	ErrDateInvalid = NewError("validation_date_invalid", "must be a valid date")
	// ErrDateOutOfRange is the error that returns in case of an invalid date.
	ErrDateOutOfRange = NewError("validation_date_out_of_range", "the date is out of range")
)

type dateRuleOptions struct {
	Layout string    `json:"layout"`
	Min    time.Time `json:"min,omitempty"`
	Max    time.Time `json:"max,omitempty"`
}

// DateRule is a validation rule that validates date/time string values.
type DateRule struct {
	dateRuleOptions
	condition     bool
	err, rangeErr Error
}

// Date returns a validation rule that checks if a string value is in a format that can be parsed into a date.
// The format of the date should be specified as the layout parameter which accepts the same value as that for time.Parse.
// For example,
//
//	validation.Date(time.ANSIC)
//	validation.Date("02 Jan 06 15:04 MST")
//	validation.Date("2006-01-02")
//
// By calling Min() and/or Max(), you can let the Date rule to check if a parsed date value is within
// the specified date range.
//
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func Date(layout string) DateRule {
	return DateRule{
		dateRuleOptions: dateRuleOptions{
			Layout: layout,
		},
		condition: true,
		err:       ErrDateInvalid,
		rangeErr:  ErrDateOutOfRange,
	}
}

func (r DateRule) RuleType() RuleType {
	return RuleTypeDate
}

func (r DateRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.dateRuleOptions)
}

// UnmarshalJSON unmarshals the JSON data into the DateRule struct.
func (r *DateRule) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.dateRuleOptions); err != nil {
		return err
	}

	r.condition = true
	r.err = ErrDateInvalid
	r.rangeErr = ErrDateOutOfRange

	return nil
}

// Error sets the error message that is used when the value being validated is not a valid date.
func (r DateRule) Error(message string) DateRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct that is used when the value being validated is not a valid date..
func (r DateRule) ErrorObject(err Error) DateRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r DateRule) When(condition bool) DateRule {
	r.condition = condition
	return r
}

// RangeError sets the error message that is used when the value being validated is out of the specified Min/Max date range.
func (r DateRule) RangeError(message string) DateRule {
	r.rangeErr = r.rangeErr.SetMessage(message)
	return r
}

// RangeErrorObject sets the error struct that is used when the value being validated is out of the specified Min/Max date range.
func (r DateRule) RangeErrorObject(err Error) DateRule {
	r.rangeErr = err
	return r
}

// Min sets the minimum date range. A zero value means skipping the minimum range validation.
func (r DateRule) Min(min time.Time) DateRule {
	r.dateRuleOptions.Min = min
	return r
}

// Max sets the maximum date range. A zero value means skipping the maximum range validation.
func (r DateRule) Max(max time.Time) DateRule {
	r.dateRuleOptions.Max = max
	return r
}

// Validate checks if the given value is a valid date.
func (r DateRule) Validate(_ context.Context, value interface{}) error {
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

	date, err := time.Parse(r.Layout, str)
	if err != nil {
		return r.err
	}

	if !r.dateRuleOptions.Min.IsZero() && r.dateRuleOptions.Min.After(date) || !r.dateRuleOptions.Max.IsZero() && date.After(r.dateRuleOptions.Max) {
		return r.rangeErr
	}

	return nil
}

func init() {
	RegisterUnmarshaller(RuleTypeDate, func(data []byte) (RuleEx, error) {
		rule := Date("")
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
