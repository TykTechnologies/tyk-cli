package utils

import (
	"fmt"
	"io"
	"log"
)

// SingleOrListIntf function converts map[string]interface{} objects into an interface slice
func MapToIntfSlice(fileMap map[string]interface{}, key string) []interface{} {
	var interfaceSlice []interface{}
	if fileMap[key] == nil {
		interfaceSlice = append(interfaceSlice, fileMap)
	} else {
		interfaceSlice = fileMap[key].([]interface{})
	}
	return interfaceSlice
}

// Print a message to an io.Writer
func PrintMessage(w io.Writer, message string) {
	fmt.Fprintln(w, message)
}

// Handle Error function prints the error if it exists
func HandleError(err error, exit bool) {
	if err != nil {
		switch exit {
		case false:
			log.Println(err)
		case true:
			log.Fatal(err)
		}
	}
}

// ReturnErr returns an error if one exists
func ReturnErr(err error) error {
	if err != nil {
		return err
	}
	return nil
}
