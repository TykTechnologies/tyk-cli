package export

import (
	"bytes"
	"fmt"
	utils "github.com/TykTechnologies/tyk-cli/utils"
	"io"
	"net/http"
	"os"
	"time"
)

func Apis(args []string) {
	if len(args) == 4 {
		authorisation, domain, port := args[0],
			utils.CheckDomain(args[1]),
			args[2]
		client := &http.Client{Timeout: 10 * time.Second}
		url := fmt.Sprintf("%s:%s/api/apis", domain, port)
		req, err := utils.HttpRequest("GET", url, authorisation, nil)
		resp, err := client.Do(req)
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
