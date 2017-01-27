package exportpkg

import (
	"bytes"
	"fmt"
	"github.com/TykTechnologies/tyk-cli/request"
	"github.com/TykTechnologies/tyk-cli/utils"
	"io"
	"net/http"
	"os"
)

// Apis is a public function for exporting APIs to a specified JSON file
func Apis(args []string) {
	if len(args) == 4 {
		call := request.New(args[0], args[1], args[2])
		req, err := call.FullRequest("GET", "/api/apis", nil)
		resp, err := call.Client.Do(req)
		outputFile := args[3]
		exportResponse(resp, err, outputFile)
	}
}

func exportResponse(resp *http.Response, err error, file string) {
	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()
		jsonString := generateJSON(resp.Body)
		absPath := utils.HandleFilePath(file)
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
