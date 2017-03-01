package utils

import (
	"fmt"
	"io"
	"os"
	"regexp"
)

// CheckDomain function checks the format of the domain that is input
func CheckDomain(inputString string) string {
	if !isProtocolPresent(inputString) {
		printMessage(os.Stdout, "Please add a protocol to your domain")
		os.Exit(-1)
	}
	return inputString
}

func printMessage(w io.Writer, message string) {
	fmt.Fprintln(w, message)
}

func isProtocolPresent(arg string) bool {
	matched, _ := regexp.MatchString("(http|https)://", arg)
	return matched
}
