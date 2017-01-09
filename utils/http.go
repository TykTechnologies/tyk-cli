package utils

import (
	"bytes"
	"fmt"
	"net/http"
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

func HttpRequest(requestType string, url string, authorisation string, payload []byte) (*http.Request, error) {
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(payload))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authorisation)

	return req, err
}
