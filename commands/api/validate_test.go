package api

import (
	"io"
	"os"
	"strings"
	"testing"
)

type ValidJSONTest struct {
	input  io.Reader
	output bool
}

func TestIsValidJSON(t *testing.T) {
	validAPI, _ := os.Open("./valid_api.json")
	headOfAPI := io.LimitReader(validAPI, 64)
	tests := []ValidJSONTest{
		{validAPI, true},
		{headOfAPI, false},
		{strings.NewReader(`{ "azul": 14 }`), false},
		{strings.NewReader(`{ "azul": 14`), false},
		{strings.NewReader(`"azul": 14}`), false},
		{strings.NewReader(`cat: persian`), false},
		{strings.NewReader(``), false},
	}
	for _, test := range tests {
		result := isValidJSON(test.input)
		if result != test.output {
			t.Fatalf(`Unexpected return value. Expected: "%v", got : "%v"`, test.output, result)
		}
	}
}
