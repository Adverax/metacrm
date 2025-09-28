package validation

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func validateMe(s string) bool {
	return s == "me"
}

func TestNewStringRule(t *testing.T) {
	v := NewStringRule("", validateMe, "abc")

	assert.NotNil(t, v.validate)
	assert.Equal(t, "", v.err.Code())
	assert.Equal(t, "abc", v.err.Message())
}

func TestNewStringRuleWithError(t *testing.T) {
	err := NewError("C", "abc")
	v := NewStringRuleWithError("", validateMe, err)

	assert.NotNil(t, v.validate)
	assert.Equal(t, err, v.err)
	assert.Equal(t, "C", v.err.Code())
	assert.Equal(t, "abc", v.err.Message())
}

func TestStringRule_Error(t *testing.T) {
	err := NewError("code", "abc")
	v := NewStringRuleWithError("", validateMe, err).Error("abc")
	assert.Equal(t, "code", v.err.Code())
	assert.Equal(t, "abc", v.err.Message())

	v2 := v.Error("correct")
	assert.Equal(t, "code", v.err.Code())
	assert.Equal(t, "correct", v2.err.Message())
	assert.Equal(t, "abc", v.err.Message())
}

func TestStringValidator_Validate(t *testing.T) {
	ctx := context.Background()
	v := NewStringRule("", validateMe, "wrong_rule").Error("wrong")

	value := "me"

	err := v.Validate(ctx, value)
	assert.Nil(t, err)

	err = v.Validate(ctx, &value)
	assert.Nil(t, err)

	value = ""

	err = v.Validate(ctx, value)
	assert.Nil(t, err)

	err = v.Validate(ctx, &value)
	assert.Nil(t, err)

	var s *string
	err = v.Validate(ctx, s)
	assert.Nil(t, err)

	err = v.Validate(ctx, "not me")
	if assert.NotNil(t, err) {
		assert.Equal(t, "wrong", err.Error())
	}

	err = v.Validate(ctx, 100)
	if assert.NotNil(t, err) {
		assert.NotEqual(t, "wrong", err.Error())
	}

	v2 := v.Error("Wrong!")
	err = v2.Validate(ctx, "not me")
	if assert.NotNil(t, err) {
		assert.Equal(t, "Wrong!", err.Error())
	}
}

func TestGetErrorFieldName(t *testing.T) {
	type A struct {
		T0 string
		T1 string `json:"t1"`
		T2 string `json:"t2,omitempty"`
		T3 string `json:",omitempty"`
		T4 string `json:"t4,x1,omitempty"`
	}
	tests := []struct {
		tag   string
		field string
		name  string
	}{
		{"t1", "T0", "T0"},
		{"t2", "T1", "t1"},
		{"t3", "T2", "t2"},
		{"t4", "T3", "T3"},
		{"t5", "T4", "t4"},
	}
	a := reflect.TypeOf(A{})
	for _, test := range tests {
		field, _ := a.FieldByName(test.field)
		assert.Equal(t, test.name, getErrorFieldName(&field), test.tag)
	}
}

func TestStringRule_ErrorObject(t *testing.T) {
	r := NewStringRule("", validateMe, "wrong_rule")

	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, "code", r.err.Code())
	assert.Equal(t, "abc", r.err.Message())
}
