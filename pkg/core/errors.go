package core

import (
	"errors"
	"strings"
)

type Errors struct {
	errors []error
}

func NewErrors(es ...error) *Errors {
	res := &Errors{
		errors: make([]error, 0),
	}
	for _, e := range es {
		res.AddError(e)
	}
	return res
}

func (that *Errors) Check(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			that.AddError(err)
		}
	}
	return that.ResError()
}

func (that *Errors) ResError() error {
	if that.IsEmpty() {
		return nil
	}

	return that
}

func (that *Errors) Error() string {
	if len(that.errors) == 0 {
		return ""
	}

	errStrings := make([]string, 0)
	for _, e := range that.errors {
		errStrings = append(errStrings, e.Error())
	}
	return strings.Join(errStrings, "\n")
}

func (that *Errors) AddError(err error) {
	that.errors = append(that.errors, err)
}

func (that *Errors) AddErrors(errs *Errors) {
	that.errors = append(that.errors, errs.errors...)
}

func (that *Errors) IsEmpty() bool {
	return len(that.errors) == 0
}

func (that *Errors) Len() int {
	return len(that.errors)
}

func (that *Errors) Is(err error) bool {
	for _, e := range that.errors {
		if errors.Is(e, err) {
			return true
		}
	}

	return false
}

func (that *Errors) Unwrap() []error {
	if that.IsEmpty() {
		return nil
	}

	return that.errors
}

func (that *Errors) Result() error {
	if !that.IsEmpty() {
		return that
	}

	return nil
}

func Check(errs ...error) error {
	es := NewErrors()
	return es.Check(errs...)
}
