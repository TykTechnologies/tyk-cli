package utils

import (
	"fmt"
	"os"
	"testing"
)

func TestHandleFilePathReturnsPath(t *testing.T) {
	expectedResult := "/home/boomy/Documents/goStuff"
	result := HandleFilePath("/home/boomy/Documents/goStuff")
	if result != expectedResult {
		t.Fatalf("Expected %s, got %s", expectedResult, result)
	}
}

func TestHandleFilePathConvertsTilde(t *testing.T) {
	expectedResult := fmt.Sprintf("%s/Documents/goStuff", os.Getenv("HOME"))
	result := HandleFilePath("~/Documents/goStuff")
	if result != expectedResult {
		t.Fatalf("Expected %s, got %s", expectedResult, result)
	}
}
