package validation

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func abcValidation(val string) bool {
	return val == "abc"
}

func TestWhen(t *testing.T) {
	errWrongAbc := NewError("wrong_abc", "wrong_abc")
	errWrongMe := NewError("wrong_me", "wrong_me")
	abcRule := NewStringRuleWithError("", abcValidation, errWrongAbc)
	validateMeRule := NewStringRuleWithError("", validateMe, errWrongMe)

	tests := []struct {
		tag       string
		condition bool
		value     interface{}
		rules     []Rule
		elseRules []Rule
		err       error
	}{
		// True condition
		{"t1.1", true, nil, []Rule{}, []Rule{}, nil},
		{"t1.2", true, "", []Rule{}, []Rule{}, nil},
		{"t1.3", true, "", []Rule{abcRule}, []Rule{}, nil},
		{"t1.4", true, 12, []Rule{Required}, []Rule{}, nil},
		{"t1.5", true, nil, []Rule{Required}, []Rule{}, ErrRequired},
		{"t1.6", true, "123", []Rule{abcRule}, []Rule{}, errWrongAbc},
		{"t1.7", true, "abc", []Rule{abcRule}, []Rule{}, nil},
		{"t1.8", true, "abc", []Rule{abcRule, abcRule}, []Rule{}, nil},
		{"t1.9", true, "abc", []Rule{abcRule, validateMeRule}, []Rule{}, errWrongMe},
		{"t1.10", true, "me", []Rule{abcRule, validateMeRule}, []Rule{}, errWrongAbc},
		{"t1.11", true, "me", []Rule{}, []Rule{abcRule}, nil},

		// False condition
		{"t2.1", false, "", []Rule{}, []Rule{}, nil},
		{"t2.2", false, "", []Rule{abcRule}, []Rule{}, nil},
		{"t2.3", false, "abc", []Rule{abcRule}, []Rule{}, nil},
		{"t2.4", false, "abc", []Rule{abcRule, abcRule}, []Rule{}, nil},
		{"t2.5", false, "abc", []Rule{abcRule, validateMeRule}, []Rule{}, nil},
		{"t2.6", false, "me", []Rule{abcRule, validateMeRule}, []Rule{}, nil},
		{"t2.7", false, "", []Rule{abcRule, validateMeRule}, []Rule{}, nil},
		{"t2.8", false, "me", []Rule{}, []Rule{abcRule, validateMeRule}, errWrongAbc},
	}

	ctx := context.Background()
	for _, test := range tests {
		err := Validate(ctx, test.value, When(test.condition, test.rules...).Else(test.elseRules...))
		assert.ErrorIs(t, err, test.err)
	}
}

type ctxKey int

const (
	contains ctxKey = iota
)

func TestWhenWithContext(t *testing.T) {
	rule := By(func(ctx context.Context, value interface{}) error {
		if !strings.Contains(value.(string), ctx.Value(contains).(string)) {
			return errors.New("unexpected value")
		}
		return nil
	})
	ctx1 := context.WithValue(context.Background(), contains, "abc")
	ctx2 := context.WithValue(context.Background(), contains, "xyz")

	tests := []struct {
		tag       string
		condition bool
		value     interface{}
		ctx       context.Context
		err       string
	}{
		// True condition
		{"t1.1", true, "abc", ctx1, ""},
		{"t1.2", true, "abc", ctx2, "unexpected value"},
		{"t1.3", true, "xyz", ctx1, "unexpected value"},
		{"t1.4", true, "xyz", ctx2, ""},

		// False condition
		{"t2.1", false, "abc", ctx1, ""},
		{"t2.2", false, "abc", ctx2, "unexpected value"},
		{"t2.3", false, "xyz", ctx1, "unexpected value"},
		{"t2.4", false, "xyz", ctx2, ""},
	}

	for _, test := range tests {
		err := Validate(test.ctx, test.value, When(test.condition, rule).Else(rule))
		assertError(t, test.err, err, test.tag)
	}
}
