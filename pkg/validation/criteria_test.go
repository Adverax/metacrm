package validation

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type person struct {
	name string
	age  int
}

func (that *person) DeclareValidationFields(declarations Declarations) {
	declarations.DeclareString("name", that.name)
	declarations.DeclareInteger("age", that.age)
}

func TestCriteria(t *testing.T) {
	rule := Criteria("this == 'test' && name == 'Alice' && age == 30")
	p := &person{name: "Alice", age: 30}
	ctx := WithThis(context.Background(), p)
	err := rule.Validate(ctx, "test")
	require.NoError(t, err)
}
