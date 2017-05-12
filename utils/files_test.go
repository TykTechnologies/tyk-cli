package utils

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestParseJSONFile(t *testing.T) {
	expectedResult := map[string]interface{}{
		"This": "is a",
		"test": map[string]interface{}{
			"this": "is a", "nest": "See?",
		},
	}
	result := ParseJSONFile("./nested_test.json")
	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("Expected %v, got %v", expectedResult, result)
	}
}

func TestParseJSONFileOutputNotFoundError(t *testing.T) {
	expectedResult := "JSON file not found"
	if os.Getenv("ERRS") == "1" {
		ParseJSONFile("./json_not_found.json")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestParseJSONFileOutputNotFoundError")
	cmd.Env = append(os.Environ(), "ERRS=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	} else {
		t.Fatalf("Expected error: %v", expectedResult)
	}
}

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

func TestMkdirPFile(t *testing.T) {
	newFiles := []string{
		"someFolder/someFile.txt",
		"someFolder/someFile2.txt",
		"some/folder/some/file.txt",
		"file.txt",
	}
	for i := range newFiles {
		MkdirPFile(newFiles[i])
		if _, err := os.Stat(newFiles[i]); os.IsNotExist(err) {
			t.Fatalf("Expected to find '%s', instead file not found", newFiles[i])
		}
		os.Remove(newFiles[i])

	}
	for i := range newFiles {
		os.RemoveAll(strings.Split(newFiles[i], "/")[0])
	}
}

func TestMkdirPFileWithNotInput(t *testing.T) {
	MkdirPFile("")
	if _, err := os.Stat(""); !os.IsNotExist(err) {
		t.Fatalf("Expected file not found, got %s", err)
	}
}
