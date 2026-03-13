package lib

import (
	"testing"
)

func TestRemoveDuplicates(t *testing.T) {
	input := []string{"a", "b", "a", "c"}
	expected := []string{"a", "b", "c"}
	result := RemoveDuplicates(input)

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("At index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestRemoveDuplicates_Empty(t *testing.T) {
	input := []string{}
	expected := []string(nil)
	result := RemoveDuplicates(input)

	if result == nil && expected == nil {
		return
	}

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
}

func TestRemoveDuplicates_NoDuplicates(t *testing.T) {
	input := []string{"a", "b"}
	expected := []string{"a", "b"}
	result := RemoveDuplicates(input)

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("At index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestRemoveEmptyString(t *testing.T) {
	input := []string{"a", "", "b", ""}
	expected := []string{"a", "b"}
	result := RemoveEmptyString(input)

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("At index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestRemoveEmptyString_AllEmpty(t *testing.T) {
	input := []string{"", ""}
	expected := []string{}
	result := RemoveEmptyString(input)

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
}
