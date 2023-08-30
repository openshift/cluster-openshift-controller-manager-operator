package controllers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestControllersSort(t *testing.T) {
	tests := []struct {
		name     string
		data     []string
		expected []string
	}{
		{
			data: []string{
				"*", "-Max", "Zea", "Ada"},
			expected: []string{
				"*", "Ada", "-Max", "Zea"},
		},
		{
			data: []string{
				"*"},
			expected: []string{
				"*"},
		},
		{
			data:     []string{},
			expected: []string{},
		},
		{
			data: []string{
				"-Max", "-Zea", "-Ada"},
			expected: []string{
				"-Ada", "-Max", "-Zea"},
		},
		{
			data: []string{
				"Joe", "Sam", "-Max", "Lee", "Rio", "Ray", "-Pax", "Ash", "Ian", "Kit", "Eve", "Tia", "Ivy", "Ava", "Ada", "Zea", "Mia"},
			expected: []string{
				"Ada", "Ash", "Ava", "Eve", "Ian", "Ivy", "Joe", "Kit", "Lee", "-Max", "Mia", "-Pax", "Ray", "Rio", "Sam", "Tia", "Zea"},
		},
		{
			data: []string{
				"Zea", "Mia", "Joe", "Sam", "-Max", "Rio", "-Pax", "Ash", "Ian", "Lee", "Kit", "Ray", "Eve", "Ivy", "Ava", "Ada", "Tia"},
			expected: []string{
				"Ada", "Ash", "Ava", "Eve", "Ian", "Ivy", "Joe", "Kit", "Lee", "-Max", "Mia", "-Pax", "Ray", "Rio", "Sam", "Tia", "Zea"},
		},
		{
			data: []string{
				"Joe", "Sam", "-Max", "Lee", "*", "Rio", "Ray", "-Pax", "Ash", "Ian", "Kit", "Eve", "Tia", "Ivy", "Ava", "Ada", "Zea", "Mia"},
			expected: []string{
				"*", "Ada", "Ash", "Ava", "Eve", "Ian", "Ivy", "Joe", "Kit", "Lee", "-Max", "Mia", "-Pax", "Ray", "Rio", "Sam", "Tia", "Zea"},
		},
	}
	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			actual := append([]string{}, tt.data...)
			controllersSort(actual).Sort()
			if !cmp.Equal(actual, tt.expected) {
				t.Log(cmp.Diff(actual, tt.expected))
				t.Fail()
			}
		})
	}
}
