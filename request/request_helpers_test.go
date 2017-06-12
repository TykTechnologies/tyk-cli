package request

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/TykTechnologies/tyk-cli/utils"
)

var newApiResponse = `{
    "Status": "OK"
    "Message": "API created"
    "Meta": "58c6cdeb0185df02ad04f43b"
}`

func TestGenerateJSON(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:3000/api/apis", nil)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(newApiResponse))
	})
	handler.ServeHTTP(recorder, req)
	expectedResult := newApiResponse
	result := GenerateJSON(recorder.Result().Body)
	if result != expectedResult {
		t.Errorf(
			"Handler returned unexpected response.\nGot:\n\t%v\nExpected:\n\t%v",
			result, expectedResult,
		)
	}
}

func TestCheckDomain(t *testing.T) {
	expectedResult := "http://www.example.com"
	result := checkDomain("http://www.example.com")
	if result != expectedResult {
		t.Fatalf("Expected %s, got %s", expectedResult, result)
	}
}

func TestCheckDomainErrorsWithBadInput(t *testing.T) {
	if os.Getenv("ERRS") == "1" {
		checkDomain("www.example.com")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckDomainErrorsWithBadInput")
	cmd.Env = append(os.Environ(), "ERRS=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	} else {
		t.Fatalf("test ran with err %v, want exit status 1", err)
	}
}

func TestPrintMessage(t *testing.T) {
	var result, expectedResult bytes.Buffer
	message := "Print me"
	utils.PrintMessage(&result, message)
	fmt.Fprint(&expectedResult, message)
	for i := range expectedResult.String() {
		if expectedResult.String()[i] != result.String()[i] {
			t.Fatalf("Expected %v, got %v", expectedResult, result)
		}
	}
}

func TestIsProtocolPresentOutputsTrueIfPresent(t *testing.T) {
	expectedResult := true
	result := isProtocolPresent("http://www.example.com")
	if result != expectedResult {
		t.Fatalf("Expected %v, got %v", expectedResult, result)
	}
}

func TestIsProtocolPresentOutputsFalseIfMissing(t *testing.T) {
	expectedResult := false
	result := isProtocolPresent("/www.example.com")
	if result != expectedResult {
		t.Fatalf("Expected %v, got %v", expectedResult, result)
	}
}

func TestOutputResponse(t *testing.T) {
	req, err := http.NewRequest("GET", "https://httpbin.org/ip", nil)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	expectedResponse := `{
  "origin": "5.153.234.114"
}
`
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	})
	handler.ServeHTTP(recorder, req)
	result := string(OutputResponse(recorder.Result()))
	if result != expectedResponse {
		t.Errorf(
			"Handler returned unexpected response.\nGot:\n%v\nExpected:\n%v",
			result,
			expectedResponse,
		)
	}
}
