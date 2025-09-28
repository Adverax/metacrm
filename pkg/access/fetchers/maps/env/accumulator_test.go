package envFetcher

import (
	"reflect"
	"testing"
)

func TestKeyPathAccumulator_Add(t *testing.T) {
	tests := []struct {
		name    string
		delim   string
		actions []struct {
			key   string
			value string
		}
		expected map[string]interface{}
	}{
		{
			name:  "single level",
			delim: ".",
			actions: []struct {
				key   string
				value string
			}{
				{"key1", "value1"},
				{"key2", "value2"},
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:  "nested keys",
			delim: ".",
			actions: []struct {
				key   string
				value string
			}{
				{"key1.subkey1", "value1"},
				{"key1.subkey2", "value2"},
			},
			expected: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": "value1",
					"subkey2": "value2",
				},
			},
		},
		{
			name:  "different delimiter",
			delim: "/",
			actions: []struct {
				key   string
				value string
			}{
				{"key1/subkey1", "value1"},
				{"key1/subkey2", "value2"},
			},
			expected: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": "value1",
					"subkey2": "value2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := NewKeyPathAccumulator(tt.delim)
			for _, action := range tt.actions {
				acc.Add(action.key, action.value)
			}
			if !reflect.DeepEqual(acc.Result(), tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, acc.Result())
			}
		})
	}
}
