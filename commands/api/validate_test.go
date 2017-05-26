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

type ValidPathTest struct {
	input  string
	output string
}

func TestHandleValidationPath(t *testing.T) {
	tests := []ValidPathTest{
		{"Object->Key[api_definition].Value->String", "[api_definition]"},
		{"Object->Key[api_definition].Value->Object->Key[org_id].Value->String", "[api_definition][org_id]"},
		{"Object->Key[api_definition].Value->Object->Key[org_id].Value->Number", "[api_definition][org_id]"},
		{"Object->Key[api_definition].Value->Object->Key[org].Value->Object->Key[id].Value->Number", "[api_definition][org][id]"},
		{"", ""},
	}
	for _, test := range tests {
		result := handleValidationPath(test.input)
		if result != test.output {
			t.Fatalf(`Unexpected return value. Expected: "%v", got : "%v"`, test.output, result)
		}
	}
}
