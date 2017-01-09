package export

import (
	"bytes"
	"fmt"
	request "github.com/TykTechnologies/tyk-cli/request"
	utils "github.com/TykTechnologies/tyk-cli/utils"
	"io"
	"net/http"
	"os"
)

func Apis(args []string) {
	if len(args) == 4 {
		call := request.New(args[0], args[1], args[2])
		url := fmt.Sprintf("%s:%s/api/apis", call.Domain, call.Port)
		req, err := call.FullRequest("GET", url, nil)
		resp, err := call.Client.Do(req)
		output_file := args[3]
		exportResponse(resp, err, output_file)
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
