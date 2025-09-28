package validation

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDependsOn(t *testing.T) {
	in := In("val1", "val2")
	rule := DependsOn("field1", &in)
	data, err := MarshalRule(&rule)
	require.NoError(t, err)
	rule2, err := UnmarshalRule(data)
	require.NoError(t, err)
	require.Equal(t, string(RuleTypeDependsOn), string(rule2.RuleType()))
}
