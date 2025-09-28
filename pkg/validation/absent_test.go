package validation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNil(t *testing.T) {
	s1 := "123"
	s2 := ""
	var time1 time.Time

	tests := []struct {
		name  string
		value any
		err   error
	}{
		{"Number", 123, ErrNil},
		{"String", "", ErrNil},
		{"Pointer to string", &s1, ErrNil},
		{"Pointer to empty string", &s2, ErrNil},
		{"Nil", nil, nil},
		{"Time", time1, ErrNil},
	}

	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := Nil
			err := r.Validate(ctx, test.value)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestEmpty(t *testing.T) {
	s1 := "123"
	s2 := ""
	time1 := time.Now()
	var time2 time.Time

	tests := []struct {
		name  string
		value interface{}
		err   error
	}{
		{"Number", 123, ErrEmpty},
		{"String", "", nil},
		{"Pointer to string", &s1, ErrEmpty},
		{"Pointer to empty string", &s2, nil},
		{"Nil", nil, nil},
		{"Time", time1, ErrEmpty},
		{"Empty time", time2, nil},
	}

	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := Empty
			err := r.Validate(ctx, test.value)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestAbsentRule_When(t *testing.T) {
	ctx := context.Background()
	r := Nil.When(false)
	err := Validate(ctx, 42, r)
	assert.Nil(t, err)

	r = Nil.When(true)
	err = Validate(ctx, 42, r)
	assert.ErrorIs(t, err, ErrNil)
}

func TestAbsentRule_Error(t *testing.T) {
	ctx := context.Background()
	r := Nil
	assert.Equal(t, "must be blank", r.Validate(ctx, "42").Error())
	assert.False(t, r.SkipNil)
	r2 := r.Error("123")
	assert.Equal(t, "must be blank", r.Validate(ctx, "42").Error())
	assert.False(t, r.SkipNil)
	assert.Equal(t, "123", r2.err.Message())
	assert.False(t, r2.SkipNil)

	r = Empty
	assert.Equal(t, "must be blank", r.Validate(ctx, "42").Error())
	assert.True(t, r.SkipNil)
	r2 = r.Error("123")
	assert.Equal(t, "must be blank", r.Validate(ctx, "42").Error())
	assert.True(t, r.SkipNil)
	assert.Equal(t, "123", r2.err.Message())
	assert.True(t, r2.SkipNil)
}

func TestAbsentRule_Error2(t *testing.T) {
	r := Nil

	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, err.Code(), r.err.Code())
	assert.Equal(t, err.Message(), r.err.Message())
	assert.NotEqual(t, err, Nil.err)
}
