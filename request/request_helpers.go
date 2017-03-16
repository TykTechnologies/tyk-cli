package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/TykTechnologies/tyk-cli/utils"
)

// Generate JSON string from io.ReadClosers like the http.Response Body
func GenerateJSON(reader io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String()
}

// MapToJSON converts map to JSON format
func MapToJSON(mapObj map[string]interface{}) (output string) {
	output = fmt.Sprintf("{\n")
	for i := range mapObj {
		output += fmt.Sprintf("  \"%v\": \"%s\"\n", i, mapObj[i])
	}
	output += fmt.Sprintf("}")
	return
}

// CheckDomain function checks the format of the domain that is input
func checkDomain(inputString string) string {
	if !isProtocolPresent(inputString) {
		utils.PrintMessage(os.Stdout, "Please add a protocol to your domain")
		os.Exit(-1)
	}
	return inputString
}

func isProtocolPresent(arg string) bool {
	matched, _ := regexp.MatchString("(http|https)://", arg)
	return matched
}

// OutputResponse function outputs the body of a response to stdout
func OutputResponse(resp *http.Response) {
	var responseMessage interface{}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(respBody, &responseMessage)
	if err != nil {
		fmt.Println(err)
	}
	msg := responseMessage.(map[string]interface{})
	utils.PrintMessage(os.Stdout, MapToJSON(msg))
}
