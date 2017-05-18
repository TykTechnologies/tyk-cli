package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ParseJSONFile function converts the contents of a JSON file into a nested map
func ParseJSONFile(inputFile string) map[string]interface{} {
	var fileObject interface{}
	file, err := ioutil.ReadFile(HandleFilePath(inputFile))
	if err != nil {
		fmt.Printf("JSON file not found\n")
		os.Exit(1)
	}
	err = json.Unmarshal([]byte(file), &fileObject)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	return fileObject.(map[string]interface{})
}

// HandleFilePath function handles special characters in file paths
func HandleFilePath(file string) string {
	homepath := fmt.Sprintf("%s/", os.Getenv("HOME"))
	replacer := strings.NewReplacer("~/", homepath)
	filtered := replacer.Replace(file)
	abs, _ := filepath.Abs(filtered)
	return abs
}

// MkdirPFile will create a file in a parent directory if file/directory
// doesn't already exist
func MkdirPFile(filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) && filePath != "" {
		path := strings.Split(filePath, "/")
		idx := len(path) - 1
		if idx > 0 {
			err = os.MkdirAll(strings.Join(path[:idx], "/"), os.ModePerm)
			HandleError(err, true)
		}
		_, err = os.Create(filePath)
		HandleError(err, true)
	}
}
