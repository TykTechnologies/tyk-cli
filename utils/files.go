package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HandleFilePath function handles special characters in file paths
func HandleFilePath(file string) string {
	homepath := fmt.Sprintf("%s/", os.Getenv("HOME"))
	replacer := strings.NewReplacer("~/", homepath)
	filtered := replacer.Replace(file)
	abs, _ := filepath.Abs(filtered)
	return abs
}
