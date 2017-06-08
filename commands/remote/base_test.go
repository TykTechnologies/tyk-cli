package remote

import (
	"bytes"
	"testing"
)

var remotes []interface{} = []interface{}{
	map[string]interface{}{
		"alias":             "default",
		"url":               "http://localhost:3000",
		"organisation_name": "Default Org",
		"org_id":            "12345678901234567890",
		"auth_token":        "1234567890abcdef",
	},
	map[string]interface{}{
		"alias":             "catChannel",
		"url":               "http://localhost:3333",
		"organisation_name": "cat Org",
		"org_id":            "12345678901234567891",
		"auth_token":        "1234567890abcdef",
	},
}

func TestList(t *testing.T) {
	var buf bytes.Buffer
	List(&buf, remotes, false)
	result := buf.String()
	expectedResult := "default\ncatChannel\n"
	if result != expectedResult {
		t.Fatalf("Error - expected %v, got %v", expectedResult, result)
	}
}

func TestListVerbose(t *testing.T) {
	var buf bytes.Buffer
	List(&buf, remotes, true)
	result := buf.String()
	expectedResult := "default - http://localhost:3000\ncatChannel - http://localhost:3333\n"
	if result != expectedResult {
		t.Fatalf("Error - expected %v, got %v", expectedResult, result)
	}
}
