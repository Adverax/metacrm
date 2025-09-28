package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultipleOf(t *testing.T) {
	ctx := context.Background()
	r := MultipleOf(10)
	assert.Equal(t, "must be multiple of 10", r.Validate(ctx, 11).Error())
	assert.Equal(t, nil, r.Validate(ctx, 20))
	assert.Equal(t, "cannot convert float32 to int64", r.Validate(ctx, float32(20)).Error())

	r2 := MultipleOf("some string ....")
	assert.Equal(t, "type not supported: string", r2.Validate(ctx, 10).Error())

	r3 := MultipleOf(uint(10))
	assert.Equal(t, "must be multiple of 10", r3.Validate(ctx, uint(11)).Error())
	assert.Equal(t, nil, r3.Validate(ctx, uint(20)))
	assert.Equal(t, "cannot convert float32 to uint64", r3.Validate(ctx, float32(20)).Error())

}

func Test_MultipleOf_Error(t *testing.T) {
	ctx := context.Background()
	r := MultipleOf(10)
	assert.Equal(t, "must be multiple of 10", r.Validate(ctx, 3).Error())

	r = r.Error("some error string ...")
	assert.Equal(t, "some error string ...", r.err.Message())
}

func TestMultipleOfRule_ErrorObject(t *testing.T) {
	r := MultipleOf(10)
	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, err.Code(), r.err.Code())
	assert.Equal(t, err.Message(), r.err.Message())
}
