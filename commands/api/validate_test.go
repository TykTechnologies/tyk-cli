package api

import (
	"testing"

	"github.com/TykTechnologies/tyk-cli/utils"
)

type ValidJSONTest struct {
	input  map[string]interface{}
	output bool
}

func TestIsValidJSON(t *testing.T) {
	validAPI := utils.ParseJSONFile("./valid_api.json")
	onlyAPIDef := validAPI["api_definition"].(map[string]interface{})
	missingAPIDef := validAPI
	delete(missingAPIDef, "api_definition")

	tests := []ValidJSONTest{
		{utils.ParseJSONFile("./valid_api.json"), true},
		{onlyAPIDef, false},
		{missingAPIDef, false},
		{map[string]interface{}(nil), false},
	}
	for _, test := range tests {
		result := isValidJSON(test.input)
		if result != test.output {
			t.Fatalf(`Unexpected return value. Expected: "%v", got : "%v"`, test.output, result)
		}
	}
}
