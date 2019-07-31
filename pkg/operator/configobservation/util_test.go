package configobservation

import (
	"reflect"
	"testing"
)

func TestConvertJSON(t *testing.T) {
	testObject := struct {
		Value0 string  `json:"value0"`
		Value1 int     `json:"value1"`
		Value2 float64 `json:"value2"`
	}{"value 0", 1, 4.2}

	for _, tt := range []struct {
		name         string
		object       interface{}
		expectsError bool
		expected     interface{}
	}{
		{
			name:     "nil",
			object:   nil,
			expected: nil,
		},
		{
			name:         "integer",
			object:       10,
			expectsError: true,
			expected:     nil,
		},
		{
			name:         "string",
			object:       "a simple string",
			expectsError: true,
			expected:     nil,
		},
		{
			name:     "a slice of strings",
			object:   []string{"a", "b", "c"},
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "a slice of integers",
			object:   []int{1, 2, 3},
			expected: []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:   "a struct",
			object: testObject,
			expected: map[string]interface{}{
				"value0": "value 0",
				"value1": float64(1),
				"value2": 4.2,
			},
		},
		{
			name: "a map",
			object: map[string]string{
				"prop0": "0",
				"prop1": "1",
				"prop2": "2",
			},
			expected: map[string]interface{}{
				"prop0": "0",
				"prop1": "1",
				"prop2": "2",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ConvertJSON(tt.object)
			if tt.expectsError && err == nil {
				t.Error("expected error, nil received instead")
			}

			if !tt.expectsError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("expected to be %v, got %v", tt.expected, res)
			}
		})
	}
}
