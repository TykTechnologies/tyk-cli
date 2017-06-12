package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ParseJSONFile function converts the contents of a JSON file into a nested map
func ParseJSONFile(inputFile string) map[string]interface{} {
	var m map[string]interface{}
	file, err := ioutil.ReadFile(HandleFilePath(inputFile))
	if err != nil {
		log.Fatal("JSON file not found")
	}
	if err := json.Unmarshal(file, &m); err != nil {
		log.Fatalf("File error: %v\n", err)
	}
	return m
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
		dir := filepath.Dir(filePath)
		if len(dir) > 1 {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		}

		_, err = os.Create(filePath)
		if err != nil {
			log.Fatal(err)
		}
	}
}
