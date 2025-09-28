package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MyInterface interface {
	Hello()
}

func TestNotNil(t *testing.T) {
	var v1 []int
	var v2 map[string]int
	var v3 *int
	var v4 interface{}
	var v5 MyInterface
	tests := []struct {
		tag   string
		value interface{}
		err   string
	}{
		{"t1", v1, "is required"},
		{"t2", v2, "is required"},
		{"t3", v3, "is required"},
		{"t4", v4, "is required"},
		{"t5", v5, "is required"},
		{"t6", "", ""},
		{"t7", 0, ""},
	}

	ctx := context.Background()
	for _, test := range tests {
		r := NotNil
		err := r.Validate(ctx, test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func Test_notNilRule_Error(t *testing.T) {
	ctx := context.Background()
	r := NotNil
	assert.Equal(t, "is required", r.Validate(ctx, nil).Error())
	r2 := r.Error("123")
	assert.Equal(t, "is required", r.Validate(ctx, nil).Error())
	assert.Equal(t, "123", r2.err.Message())
}

func TestNotNilRule_ErrorObject(t *testing.T) {
	r := NotNil

	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, err.Code(), r.err.Code())
	assert.Equal(t, err.Message(), r.err.Message())
	assert.NotEqual(t, err, NotNil.err)
}
