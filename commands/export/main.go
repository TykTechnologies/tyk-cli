package export

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func Apis(args []string) {
	if len(args) == 4 {
		authorisation := args[0]
		domain := checkDomain(args[1])
		port := args[2]
		client := &http.Client{Timeout: 10 * time.Second}
		url := fmt.Sprintf("%s:%s/api/apis", domain, port)
		req, err := httpRequest("GET", url, authorisation)
		resp, err := client.Do(req)
		output_file := args[3]
		exportResponse(resp, err, output_file)
	}
}

func checkDomain(inputString string) string {
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

func httpRequest(requestType string, url string, authorisation string) (*http.Request, error) {
	req, err := http.NewRequest(requestType, url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authorisation)

	return req, err
}

type ExportJson struct {
	Body string
}

func exportResponse(resp *http.Response, err error, file string) {
	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()
		jsonString := generateJSON(resp.Body)
		absPath := handleFilePath(file)
		f, err := os.Create(absPath)
		if err != nil {
			return
		}
		defer f.Close()
		f.WriteString(jsonString)
	}
}

func generateJSON(reader io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String()
}

func handleFilePath(file string) string {
	replacer := strings.NewReplacer("~", os.Getenv("HOME"))
	filtered := replacer.Replace(file)
	abs, _ := filepath.Abs(filtered)
	return abs
}
