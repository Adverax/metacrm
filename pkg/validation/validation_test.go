package validation

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	slice := []String123{String123("abc"), String123("123"), String123("xyz")}
	ctxSlice := []Model4{{A: "abc"}, {A: "def"}}
	mp := map[string]String123{"c": String123("abc"), "b": String123("123"), "a": String123("xyz")}
	var (
		ptr *string
	)

	tests := []struct {
		tag   string
		value interface{}
		err   string
	}{
		{"t1", 123, ""},
		{"t2", String123("123"), ""},
		{"t3", String123("abc"), `{"errors": ["error 123"]}`},
		{"t4", []String123{}, ""},

		{"t5", slice, `{"children": {"0": {"errors": ["error 123"]}, "2": {"errors": ["error 123"]}}}`},
		{"t6", &slice, `{"children": {"0": {"errors": ["error 123"]}, "2": {"errors": ["error 123"]}}}`},
		{"t7", ctxSlice, `{"children": {"1": {"children": {"A": {"errors": ["error abc"]}}}}}`},
		{"t8", mp, `{"children": {"a": {"errors": ["error 123"]}, "c": {"errors": ["error 123"]}}}`},
		{"t9", &mp, `{"children": {"a" : {"errors": [ "error 123" ]}, "c" : {"errors": [ "error 123" ]}}}`},

		{"t10", map[string]String123{}, ""},
		{"t11", ptr, ""},
	}
	ctx := context.Background()
	for _, test := range tests {
		err := Validate(ctx, test.value)
		assertJson(t, test.err, err, test.tag)
	}

	// with rules
	err := Validate(ctx, "123", &validateAbc{}, &validateXyz{})
	assert.ErrorIs(t, err, errABC)
	err = Validate(ctx, "abc", &validateAbc{}, &validateXyz{})
	assert.ErrorIs(t, err, errXYZ)
	err = Validate(ctx, "abcxyz", &validateAbc{}, &validateXyz{})
	assert.NoError(t, err)

	err = Validate(ctx, "123", &validateAbc{}, Skip, &validateXyz{})
	assert.ErrorIs(t, err, errABC)
	err = Validate(ctx, "abc", &validateAbc{}, Skip, &validateXyz{})
	assert.NoError(t, err)

	err = Validate(ctx, "123", &validateAbc{}, Skip.When(true), &validateXyz{})
	assert.ErrorIs(t, err, errABC)
	err = Validate(ctx, "abc", &validateAbc{}, Skip.When(true), &validateXyz{})
	assert.NoError(t, err)

	err = Validate(ctx, "123", &validateAbc{}, Skip.When(false), &validateXyz{})
	assert.ErrorIs(t, err, errABC)
	err = Validate(ctx, "abc", &validateAbc{}, Skip.When(false), &validateXyz{})
	assert.ErrorIs(t, err, errXYZ)
}

func stringEqual(str string) RuleFunc {
	return func(ctx context.Context, value interface{}) error {
		s, _ := value.(string)
		if s != str {
			return errors.New("unexpected string")
		}
		return nil
	}
}

func TestBy(t *testing.T) {
	ctx := context.Background()

	errMustBeAbc := errors.New("must be abc")
	abcRule := By(func(ctx context.Context, value interface{}) error {
		s, _ := value.(string)
		if s != "abc" {
			return errMustBeAbc
		}
		return nil
	})

	assert.Nil(t, Validate(ctx, "abc", abcRule))
	err := Validate(ctx, "xyz", abcRule)
	if assert.NotNil(t, err) {
		assert.ErrorIs(t, err, errMustBeAbc)
	}

	xyzRule := By(stringEqual("xyz"))
	assert.Nil(t, Validate(ctx, "xyz", xyzRule))
	assert.NotNil(t, Validate(ctx, "abc", xyzRule))
	assert.Nil(t, Validate(context.Background(), "xyz", xyzRule))
	assert.NotNil(t, Validate(context.Background(), "abc", xyzRule))
}

func Test_skipRule_Validate(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, Skip.Validate(ctx, 100))
}

func assertJson(t *testing.T, expected string, err error, tag string) {
	if expected == "" {
		assert.NoError(t, err, tag)
	} else {
		data, e := json.Marshal(err)
		require.NoError(t, e, tag)
		assert.JSONEq(t, expected, string(data), tag)
	}
}

func assertError(t *testing.T, expected string, err error, tag string) {
	if expected == "" {
		assert.NoError(t, err, tag)
	} else {
		assert.EqualError(t, err, expected, tag)
	}
}

var errABC = NewError("error abc", "error abc")

type validateAbc struct{}

func (v *validateAbc) Validate(_ context.Context, obj interface{}) error {
	if !strings.Contains(obj.(string), "abc") {
		return errABC
	}
	return nil
}

var errXYZ = NewError("error xyz", "error xyz")

type validateXyz struct{}

func (v *validateXyz) Validate(ctx context.Context, obj interface{}) error {
	if !strings.Contains(obj.(string), "xyz") {
		return errXYZ
	}
	return nil
}

type validateInternalError struct{}

func (v *validateInternalError) Validate(_ context.Context, obj interface{}) error {
	if strings.Contains(obj.(string), "internal") {
		return NewInternalError(errors.New("error internal"))
	}
	return nil
}

type Model1 struct {
	A string
	B string
	c string
	D *string
	E String123
	F *String123
	G string `json:"g"`
	H []string
	I map[string]string
}

var err123 = NewError("error 123", "error 123")

type String123 string

func (s String123) Validate(ctx context.Context) error {
	if !strings.Contains(string(s), "123") {
		return err123
	}
	return nil
}

type Model2 struct {
	Model3
	M3 Model3
	B  string
}

type Model3 struct {
	A string
}

func (m Model3) Validate(ctx context.Context) error {
	return ValidateStruct(ctx, &m,
		Field(&m.A, &validateAbc{}),
	)
}

type Model4 struct {
	A string
}

func (m Model4) Validate(ctx context.Context) error {
	return ValidateStruct(ctx, &m,
		Field(&m.A, &validateAbc{}),
	)
}

type Model5 struct {
	Model4
	M4 Model4
	B  string
}
