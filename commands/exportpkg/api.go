package exportpkg

import (
	"fmt"
	"net/http"
	"os"

	"github.com/TykTechnologies/tyk-cli/request"
	"github.com/TykTechnologies/tyk-cli/utils"
)

// Apis is a public function for exporting APIs to a specified JSON file
func APIs(args []string) {
	var (
		err  error
		req  *http.Request
		call *request.Request
	)
	switch len(args) {
	case 4:
		call = request.New(args[0], args[1], args[2])
		req, err = call.FullRequest("GET", "/api/apis", nil)
	case 5:
		call = request.New(args[1], args[2], args[3])
		path := fmt.Sprintf("/api/apis/%s", args[0])
		req, err = call.FullRequest("GET", path, nil)
	}
	utils.HandleError(err, true)
	resp, err := call.Client.Do(req)
	utils.HandleError(err, false)
	if resp.StatusCode != 200 {
		fmt.Println(resp.Status)
		os.Exit(-1)
	}
	outputFile := args[4]
	utils.HandleError(err, false)
	exportResponse(resp, outputFile)
}

func exportResponse(resp *http.Response, file string) {
	defer resp.Body.Close()
	jsonString := request.GenerateJSON(resp.Body)
	absPath := utils.HandleFilePath(file)
	f, err := os.Create(absPath)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(jsonString)
}
