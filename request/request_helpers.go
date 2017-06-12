package request

import (
	"bytes"
	"encoding/json"
	"io"
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
func OutputResponse(resp *http.Response) []byte {
	defer resp.Body.Close()
	var respBody map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	utils.HandleError(err, false)
	output, err := json.MarshalIndent(respBody, "", "  ")
	utils.HandleError(err, true)
	return append(output, []byte("\n")[0])
}
