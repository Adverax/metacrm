package validation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequired(t *testing.T) {
	s1 := "123"
	s2 := ""
	var time1 time.Time
	tests := []struct {
		tag   string
		value interface{}
		err   string
	}{
		{"t1", 123, ""},
		{"t2", "", "cannot be blank"},
		{"t3", &s1, ""},
		{"t4", &s2, "cannot be blank"},
		{"t5", nil, "cannot be blank"},
		{"t6", time1, "cannot be blank"},
	}

	ctx := context.Background()
	for _, test := range tests {
		r := Required
		err := r.Validate(ctx, test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestRequiredRule_When(t *testing.T) {
	ctx := context.Background()
	r := Required.When(false)
	err := Validate(ctx, nil, r)
	assert.Nil(t, err)

	r = Required.When(true)
	err = r.Validate(ctx, nil)
	assert.Equal(t, ErrRequired, err)
}

func TestNilOrNotEmpty(t *testing.T) {
	s1 := "123"
	s2 := ""
	tests := []struct {
		tag   string
		value interface{}
		err   string
	}{
		{"t1", 123, ""},
		{"t2", "", "cannot be blank"},
		{"t3", &s1, ""},
		{"t4", &s2, "cannot be blank"},
		{"t5", nil, ""},
	}

	ctx := context.Background()
	for _, test := range tests {
		r := NilOrNotEmpty
		err := r.Validate(ctx, test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func Test_requiredRule_Error(t *testing.T) {
	ctx := context.Background()
	r := Required
	assert.Equal(t, "cannot be blank", r.Validate(ctx, nil).Error())
	assert.False(t, r.SkipNil)
	r2 := r.Error("123")
	assert.Equal(t, "cannot be blank", r.Validate(ctx, nil).Error())
	assert.False(t, r.SkipNil)
	assert.Equal(t, "123", r2.err.Message())
	assert.False(t, r2.SkipNil)

	r = NilOrNotEmpty
	assert.Equal(t, "cannot be blank", r.Validate(ctx, "").Error())
	assert.True(t, r.SkipNil)
	r2 = r.Error("123")
	assert.Equal(t, "cannot be blank", r.Validate(ctx, "").Error())
	assert.True(t, r.SkipNil)
	assert.Equal(t, "123", r2.err.Message())
	assert.True(t, r2.SkipNil)
}

func TestRequiredRule_Error(t *testing.T) {
	r := Required

	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, err.Code(), r.err.Code())
	assert.Equal(t, err.Message(), r.err.Message())
	assert.NotEqual(t, err, Required.err)
}
