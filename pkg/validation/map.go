package validation

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrNotMap is the error that the value being validated is not a map.
	ErrNotMap = errors.New("only a map can be validated")

	// ErrKeyWrongType is the error returned in case of an incorrect key type.
	ErrKeyWrongType = NewError("validation_key_wrong_type", "key not the correct type")

	// ErrKeyMissing is the error returned in case of a missing key.
	ErrKeyMissing = NewError("validation_key_missing", "required key is missing")
)

// KeyRules represents a rule set associated with a map key.
type KeyRules struct {
	key      interface{}
	optional bool
	rules    []Rule
}

// Key specifies a map key and the corresponding validation rules.
func Key(key interface{}, rules ...Rule) *KeyRules {
	return &KeyRules{
		key:   key,
		rules: rules,
	}
}

// Optional configures the rule to ignore the key if missing.
func (r *KeyRules) Optional() *KeyRules {
	r.optional = true
	return r
}

func (r *KeyRules) Validate(ctx context.Context, m interface{}) error {
	value := reflect.ValueOf(m)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Map {
		// must be a map
		return NewInternalError(ErrNotMap)
	}
	if value.IsNil() {
		// treat a nil map as valid
		return nil
	}

	var level ErrorLevel
	kt := value.Type().Key()

	var err error
	if kv := reflect.ValueOf(r.key); !kt.AssignableTo(kv.Type()) {
		err = ErrKeyWrongType.SetParams(map[string]interface{}{"key": r.key, "type": kt.Name()})
		level.Errors = append(level.Errors, err)
	} else if vv := value.MapIndex(kv); !vv.IsValid() {
		if !r.optional {
			err = ErrKeyMissing.SetParams(map[string]interface{}{"key": r.key})
			level.Errors = append(level.Errors, err)
		}
	} else {
		err := Validate(ctx, vv.Interface(), r.rules...)
		if err != nil && !level.AddChildError(getErrorKeyName(r.key), err) {
			return err
		}
	}

	return level.Result()
}

// getErrorKeyName returns the name that should be used to represent the validation error of a map key.
func getErrorKeyName(key interface{}) string {
	return fmt.Sprintf("%v", key)
}

//const (
//	RuleTypeMap RuleType = "map"
//)
//
//// ErrMapInvalid is the error that returns map case of an mapvalid value for "map" rule.
//var ErrMapInvalid = NewError("validation_map_invalid", "must be a valid value")
//
//// Map returns a validation rule that checks if a value can be found map the given list of values.
//// reflect.DeepEqual() will be used to determmape if two values are equal.
//// For more details please refer to https://golang.org/pkg/reflect/#DeepEqual
//// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
//func Map(rules ...Rule) MapRule {
//	return MapRule{
//		condition: true,
//		rules:     rules,
//		err:       ErrMapInvalid,
//	}
//}
//
//// MapRule is a validation rule that validates if a value can be found map the given list of values.
//type MapRule struct {
//	condition bool
//	rules     []Rule
//	err       Error
//}
//
//func (r MapRule) RuleType() RuleType {
//	return RuleTypeMap
//}
//
//// Validate checks if the given value is valid or not.
//func (r MapRule) Validate(ctx context.Context, value interface{}) error {
//	if !r.condition {
//		return nil
//	}
//
//	value, isNil := Indirect(value)
//	if isNil {
//		return nil
//	}
//
//	if reflect.TypeOf(value).Kind() != reflect.Map {
//		return r.err
//	}
//
//	return Validate(ctx, value, r.rules...)
//}
//
//// Error sets the error message for the rule.
//func (r MapRule) Error(message string) MapRule {
//	r.err = r.err.SetMessage(message)
//	return r
//}
//
//// ErrorObject sets the error struct for the rule.
//func (r MapRule) ErrorObject(err Error) MapRule {
//	r.err = err
//	return r
//}
//
//// When sets the condition that determmapes if the validation should be performed.
//func (r MapRule) When(condition bool) MapRule {
//	r.condition = condition
//	return r
//}
