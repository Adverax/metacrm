package configs

import (
	"reflect"
	"testing"
)

func TestOverride(t *testing.T) {
	tests := []struct {
		a, b, expected map[string]interface{}
	}{
		{
			a: map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"subkey1": "subvalue1",
				},
			},
			b: map[string]interface{}{
				"key2": map[string]interface{}{
					"subkey1": "newsubvalue1",
					"subkey2": "subvalue2",
				},
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"subkey1": "newsubvalue1",
					"subkey2": "subvalue2",
				},
				"key3": "value3",
			},
		},
		{
			a: map[string]interface{}{
				"key1": []interface{}{1, 2, 3},
			},
			b: map[string]interface{}{
				"key1": []interface{}{4, 5, 6},
			},
			expected: map[string]interface{}{
				"key1": []interface{}{4, 5, 6},
			},
		},
	}

	for _, tt := range tests {
		override(tt.a, tt.b)
		if !reflect.DeepEqual(tt.a, tt.expected) {
			t.Errorf("override() = %v, want %v", tt.a, tt.expected)
		}
	}
}
