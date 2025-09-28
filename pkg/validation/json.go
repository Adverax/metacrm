package validation

import (
	"context"
	"encoding/json"
)

var (
	// ErrJsonInvalid is the error that returns in case of an invalid json.
	ErrJsonInvalid = NewError("validation_json_invalid", "must be a valid json")
	// ErrJsonInvalidMarshall is the error that returns in case of an error during marshalling.
	ErrJsonInvalidMarshall = NewError("validation_json_invalid_marshall", "error marshalling json")
	// ErrJsonInvalidUnmarshall is the error that returns in case of an error during unmarshalling.
	ErrJsonInvalidUnmarshall = NewError("validation_json_invalid_unmarshall", "error unmarshalling json")
)

// JsonSchema is an interface that defines a method for validating JSON values.
// It may be implemented by any type that can validate JSON data.
// For example: github.com/santhosh-tekuri/jsonschema/v5
type JsonSchema interface {
	Validate(ctx context.Context, value interface{}) error
}

// JsonRule is a validation rule that valijsons json/time string values.
type JsonRule struct {
	condition bool
	schema    JsonSchema
	layout    string
	err       Error
}

func Json(schema JsonSchema) JsonRule {
	return JsonRule{
		condition: true,
		schema:    schema,
		err:       ErrJsonInvalid,
	}
}

// Error sets the error message that is used when the value being valijsond is not a valid json.
func (r JsonRule) Error(message string) JsonRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct that is used when the value being valijsond is not a valid json..
func (r JsonRule) ErrorObject(err Error) JsonRule {
	r.err = err
	return r
}

// When sets the condition that determines if the validation should be performed.
func (r JsonRule) When(condition bool) JsonRule {
	r.condition = condition
	return r
}

// Validate checks if the given value is a valid json.
func (r JsonRule) Validate(ctx context.Context, value interface{}) error {
	if !r.condition {
		return nil
	}

	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return ErrJsonInvalidMarshall
	}

	var val interface{}
	if err := json.Unmarshal(data, &val); err != nil {
		return ErrJsonInvalidUnmarshall
	}

	err = r.schema.Validate(ctx, val)
	if err != nil {
		return r.err.SetMessage(err.Error())
	}

	return nil
}
