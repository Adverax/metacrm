package validation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

type (
	// Error interface represents an validation error
	Error interface {
		Error() string
		Code() string
		Message() string
		Cause() error
		SetCode(string) Error
		SetMessage(string) Error
		SetCause(error) Error
		Params() map[string]interface{}
		SetParams(map[string]interface{}) Error
		AddParam(name string, value interface{}) Error
	}

	// ErrorObject is the default validation error
	// that implements the Error interface.
	ErrorObject struct {
		code    string
		message string
		cause   error
		params  map[string]interface{}
	}

	// Errors represents the validation errors that are indexed by struct field names, map or slice keys.
	// values are Error or Errors (for map, slice and array error value is Errors).
	Errors map[string]error

	// ErrorList represents a list of errors.
	ErrorList []error

	ErrorLevel struct {
		Errors   ErrorList
		Children Errors
	}

	// InternalError represents an error that should NOT be treated as a validation error.
	InternalError interface {
		error
		InternalError() error
	}

	internalError struct {
		error
	}
)

// NewInternalError wraps a given error into an InternalError.
func NewInternalError(err error) InternalError {
	return internalError{error: err}
}

// InternalError returns the actual error that it wraps around.
func (e internalError) InternalError() error {
	return e.error
}

// SetCode set the error's translation code.
func (e ErrorObject) SetCode(code string) Error {
	e.code = code
	return e
}

// Code get the error's translation code.
func (e ErrorObject) Code() string {
	return e.code
}

// Cause returns the cause of the error, if any.
func (e ErrorObject) Cause() error {
	return e.cause
}

// SetCause sets the cause of the error.
func (e ErrorObject) SetCause(cause error) Error {
	e.cause = cause
	if cause == nil {
		if e.params != nil {
			delete(e.params, "cause")
		}
		return e
	}
	return e.AddParam("cause", cause.Error())
}

// SetParams set the error's params.
func (e ErrorObject) SetParams(params map[string]interface{}) Error {
	e.params = params
	return e
}

// AddParam add parameter to the error's parameters.
func (e ErrorObject) AddParam(name string, value interface{}) Error {
	if e.params == nil {
		e.params = make(map[string]interface{})
	}

	e.params[name] = value
	return e
}

// Params returns the error's params.
func (e ErrorObject) Params() map[string]interface{} {
	return e.params
}

// SetMessage set the error's message.
func (e ErrorObject) SetMessage(message string) Error {
	e.message = message
	return e
}

// Message return the error's message.
func (e ErrorObject) Message() string {
	return e.message
}

// Error returns the error message.
func (e ErrorObject) Error() string {
	if len(e.params) == 0 {
		return e.message
	}

	res := bytes.Buffer{}
	_ = template.Must(template.New("err").Parse(e.message)).Execute(&res, e.params)

	return res.String()
}

// Is implements the errors.Is interface to check if the error matches a target error.
func (e ErrorObject) Is(target error) bool {
	if target == nil {
		return false
	}

	if e2, ok := target.(ErrorObject); ok {
		match := e.code == e2.code && e.message == e2.message && reflect.DeepEqual(e.params, e2.params)
		if match {
			return true
		}
		if e.cause != nil {
			if errors.Is(e.cause, target) {
				return true
			}
		}
	}

	return false
}

func (e ErrorObject) Unwrap() []error {
	if e.cause == nil {
		return nil
	}
	return []error{e.cause}
}

// Error returns the error string of Errors.
func (es Errors) Error() string {
	if len(es) == 0 {
		return ""
	}

	keys := make([]string, len(es))
	i := 0
	for key := range es {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	var s strings.Builder
	for i, key := range keys {
		if i > 0 {
			s.WriteString("; ")
		}
		if errs, ok := es[key].(*ErrorLevel); ok {
			_, _ = fmt.Fprintf(&s, "%v: (%v)", key, errs)
		} else {
			_, _ = fmt.Fprintf(&s, "%v: %v", key, es[key].Error())
		}
	}
	s.WriteString(".")
	return s.String()
}

// MarshalJSON converts the Errors into a valid JSON.
func (es Errors) MarshalJSON() ([]byte, error) {
	errs := map[string]interface{}{}
	for key, err := range es {
		if ms, ok := err.(json.Marshaler); ok {
			errs[key] = ms
		} else {
			errs[key] = err.Error()
		}
	}
	return json.Marshal(errs)
}

// Is implements the errors.Is interface to check if the error matches a target error.
func (es Errors) Is(target error) bool {
	for _, err := range es {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// Unwrap returns a slice of errors contained in Errors.
func (es Errors) Unwrap() []error {
	if len(es) == 0 {
		return nil
	}
	errs := make([]error, 0, len(es))
	for _, err := range es {
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Filter removes all nils from Errors and returns back the updated Errors as an error.
// If the length of Errors becomes 0, it will return nil.
func (es Errors) Filter() error {
	for key, value := range es {
		if value == nil {
			delete(es, key)
		}
	}
	if len(es) == 0 {
		return nil
	}
	return es
}

// Is implements the errors.Is interface to check if the error matches a target error.
func (errs ErrorList) Is(target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// Unwrap returns a slice of errors contained in ErrorList.
func (errs ErrorList) Unwrap() []error {
	if len(errs) == 0 {
		return nil
	}
	errors := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (errs ErrorList) Error() string {
	if len(errs) == 0 {
		return "[]"
	}

	list := make([]string, len(errs))
	for i, err := range errs {
		if err != nil {
			list[i] = err.Error()
		} else {
			list[i] = "nil"
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(list, "; "))
}

func (errs ErrorList) MarshalJSON() ([]byte, error) {
	if len(errs) == 0 {
		return []byte("[]"), nil
	}

	list := make([]interface{}, len(errs))
	for i, err := range errs {
		if err != nil {
			if ms, ok := err.(json.Marshaler); ok {
				list[i] = ms
			} else {
				list[i] = err.Error()
			}
		}
	}

	return json.Marshal(list)
}

func NewErrorLevel() *ErrorLevel {
	return &ErrorLevel{
		Children: make(Errors),
	}
}

func (l *ErrorLevel) Error() string {
	var s strings.Builder
	if len(l.Errors) > 0 {
		s.WriteString("Errors: [")
		for i, err := range l.Errors {
			if i > 0 {
				s.WriteString("; ")
			}
			s.WriteString(err.Error())
		}
		s.WriteString("]")
	}

	if len(l.Children) > 0 {
		if s.Len() > 0 {
			s.WriteString("; ")
		}
		s.WriteString("Children: {")
		first := true
		for key, err := range l.Children {
			if first {
				first = false
			} else {
				s.WriteString("; ")
			}
			s.WriteString(fmt.Sprintf("%s: %s", key, err.Error()))
		}
		s.WriteString("}")
	}

	return s.String()
}

func (l *ErrorLevel) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}

	if len(l.Errors) > 0 {
		data["errors"] = l.Errors
	}

	if len(l.Children) > 0 {
		data["children"] = l.Children
	}

	return json.Marshal(data)
}

func (l *ErrorLevel) Is(target error) bool {
	if target == nil {
		return false
	}

	if x, ok := target.(*ErrorLevel); ok {
		return reflect.DeepEqual(l, x)
	}

	if len(l.Errors) != 0 && l.Errors.Is(target) {
		return true
	}

	if len(l.Children) != 0 && l.Children.Is(target) {
		return true
	}

	return false
}

func (l *ErrorLevel) Unwrap() []error {
	if len(l.Errors) == 0 && len(l.Children) == 0 {
		return nil
	}

	errs := make([]error, 0, len(l.Errors)+len(l.Children))
	for _, err := range l.Errors {
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, err := range l.Children {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (l *ErrorLevel) IsEmpty() bool {
	return len(l.Errors) == 0 && len(l.Children) == 0
}

func (l *ErrorLevel) AddError(err error) bool {
	if err == nil {
		return false
	}

	if !IsValidationError(err) {
		return false
	}

	if e, ok := err.(*ErrorLevel); ok {
		l.Errors = append(l.Errors, e.Errors...)
		for key, child := range e.Children {
			if existing, found := l.Children[key]; found {
				if existingLevel, ok := existing.(*ErrorLevel); ok {
					existingLevel.AddError(child)
				} else {
					l.putChildError(key, &ErrorLevel{Errors: ErrorList{child}})
				}
			} else {
				l.putChildError(key, child)
			}
		}
	} else if e, ok := err.(Errors); ok {
		for key, child := range e {
			if existing, found := l.Children[key]; found {
				if existingLevel, ok := existing.(*ErrorLevel); ok {
					existingLevel.AddError(child)
				} else {
					l.putChildError(key, &ErrorLevel{Errors: ErrorList{child}})
				}
			} else {
				l.putChildError(key, child)
			}
		}
	} else {
		l.Errors = append(l.Errors, err)
	}

	return true
}

func (l *ErrorLevel) AddChildError(key string, err error) bool {
	if err == nil {
		return false
	}
	if !IsValidationError(err) {
		return false
	}

	lvl := EnsureLevel(err)
	if l.Children != nil {
		if child, exists := l.Children[key]; exists {
			lvl.AddError(child)
		}
	}
	l.putChildError(key, lvl)
	return true
}

func (l *ErrorLevel) putChildError(key string, err error) {
	if l.Children == nil {
		l.Children = make(Errors)
	}

	l.Children[key] = err
}

func (l *ErrorLevel) Merge(other *ErrorLevel) {
	if other == nil {
		return
	}

	l.Errors = append(l.Errors, other.Errors...)

	for key, child := range other.Children {
		if existing, found := l.Children[key]; found {
			if existingLevel, ok := existing.(*ErrorLevel); ok {
				if childLevel, ok := child.(*ErrorLevel); ok {
					existingLevel.Merge(childLevel)
				} else {
					existingLevel.Errors = append(existingLevel.Errors, child)
				}
			} else {
				if childLevel, ok := child.(*ErrorLevel); ok {
					childLvl := &ErrorLevel{
						Errors:   ErrorList{existing},
						Children: make(Errors),
					}
					l.Children[key] = childLvl
					childLvl.Merge(childLevel)
				} else {
					l.Children[key] = &ErrorLevel{
						Errors:   ErrorList{existing, child},
						Children: make(Errors),
					}
				}
			}
		} else {
			l.Children[key] = child
		}
	}
}

func (l *ErrorLevel) UpdateChildrenNamesByModelTags(model interface{}) {
	l.UpdateChildrenNamesByTypeTags(reflect.TypeOf(model))
}

func (l *ErrorLevel) UpdateChildrenNamesByTypeTags(typ reflect.Type) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	var toDelete []string
	for key, child := range l.Children {
		switch typ.Kind() {
		case reflect.Struct:
			if sf, found := typ.FieldByNameFunc(func(name string) bool {
				return strings.ToLower(name) == strings.ToLower(key)
			}); found {
				if tag := sf.Tag.Get("json"); tag != "" && tag != "-" {
					if jsonName := strings.SplitN(tag, ",", 2)[0]; jsonName != "" && jsonName != key {
						l.Children[jsonName] = child
						toDelete = append(toDelete, key)
					}
				}
				var childLevel *ErrorLevel
				if errors.As(child, &childLevel) {
					childLevel.UpdateChildrenNamesByTypeTags(sf.Type)
				}
			}
		case reflect.Map:
			if typ.Key().Kind() == reflect.String {
				var childLevel *ErrorLevel
				if errors.As(child, &childLevel) {
					childLevel.UpdateChildrenNamesByTypeTags(typ.Elem())
				}
			}
		case reflect.Slice, reflect.Array:
			if _, err := strconv.Atoi(key); err == nil {
				var childLevel *ErrorLevel
				if errors.As(child, &childLevel) {
					childLevel.UpdateChildrenNamesByTypeTags(typ.Elem())
				}
			}
		default:
		}
	}

	for _, key := range toDelete {
		delete(l.Children, key)
	}
}

func (l *ErrorLevel) Result() error {
	if l.IsEmpty() {
		return nil
	}

	return l
}

// NewError create new validation error.
func NewError(code, message string) Error {
	return ErrorObject{
		code:    code,
		message: message,
	}
}

type ErrorDictionary map[string]ErrorList

func (d ErrorDictionary) IsEmpty() bool {
	return len(d) == 0
}

func (d ErrorDictionary) Add(key string, err error) {
	if d[key] == nil {
		d[key] = make(ErrorList, 0, 1)
	}

	if e, ok := err.(ErrorList); ok {
		d[key] = append(d[key], e...)
	} else {
		d[key] = append(d[key], err)
	}
}

func (d ErrorDictionary) ToErrors() Errors {
	if len(d) == 0 {
		return nil
	}

	es := make(Errors, len(d))
	for key, errs := range d {
		if len(errs) == 0 {
			continue
		}
		es[key] = errs
	}

	return es
}

// Assert that our ErrorObject implements the Error interface.
var _ Error = ErrorObject{}
