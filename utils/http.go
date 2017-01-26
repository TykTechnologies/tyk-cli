package utils

import (
	"fmt"
	"os"
	"regexp"
)

// CheckDomain function checks the format of the domain that is input
func CheckDomain(inputString string) string {
	if !isProtocolPresent(inputString) {
		fmt.Println("Please add a protocol to your domain")
		os.Exit(-1)
	}
	return inputString
}

func isProtocolPresent(arg string) bool {
	matched, _ := regexp.MatchString("(http|https)://", arg)
	return matched
}
