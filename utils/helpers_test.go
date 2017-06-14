package utils

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestMapToIntfSlice(t *testing.T) {
	expectedResult := []interface{}{
		map[string]interface{}{
			"id":   1,
			"name": "Lucrezia Purrgia",
		},
		map[string]interface{}{
			"id":   2,
			"name": "Snowball",
		},
	}
	petshop := map[string]interface{}{
		"cats": expectedResult,
		"dogs": "n/a",
	}
	result := MapToIntfSlice(petshop, "cats")
	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("Expected %v, got %v", expectedResult, result)
	}
}

func TestMapToIntfSliceSingle(t *testing.T) {
	cat := map[string]interface{}{
		"id":   1,
		"name": "Lucrezia Purrgia",
	}
	expectedResult := []interface{}{cat}
	result := MapToIntfSlice(cat, "cats")
	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("Expected %v, got %v", expectedResult, result)
	}
}

func TestPrintMessage(t *testing.T) {
	var resultBuffer bytes.Buffer
	message := "Cats"
	PrintMessage(&resultBuffer, message)
	expectedResult := strings.TrimSpace(message)
	result := strings.TrimSpace(resultBuffer.String())
	if result != expectedResult {
		t.Fatalf("Expected: %v, got: %v", expectedResult, result)
	}
}

type ContainsTest struct {
	inputArr []string
	inputS   string
	output   bool
}

func TestContains(t *testing.T) {
	tests := []ContainsTest{
		{[]string{"cat", "dog", "hamster", "rabbit"}, "cat", true},
		{[]string{"cat", "dog", "hamster", "rabbit"}, "pony", false},
		{[]string{}, "cat", false},
	}

	for _, test := range tests {
		result := Contains(test.inputArr, test.inputS)
		if result != test.output {
			t.Fatalf(`Unexpected return value. Expected: "%v", got : "%v"`, test.output, result)

		}

	}
}
