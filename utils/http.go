package utils

import (
	"fmt"
	"os"
	"regexp"
)

func CheckDomain(inputString string) string {
	if !IsProtocolPresent(inputString) {
		fmt.Println("Please add a protocol to your domain")
		os.Exit(-1)
	}
	return inputString
}

func IsProtocolPresent(arg string) bool {
	matched, _ := regexp.MatchString("(http|https)://", arg)
	return matched
}
