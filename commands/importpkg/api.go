package importpkg

import (
	"encoding/json"
	"fmt"

	"github.com/TykTechnologies/tyk-cli/request"
	"github.com/TykTechnologies/tyk-cli/utils"
)

// Apis is a public function for importing APIs
func APIs(args []string) {
	if len(args) == 4 {
		call := request.New(args[0], args[1], args[2])
		inputFile := args[3]
		fileMap := utils.ParseJSONFile(inputFile)
		apis := utils.MapToIntfSlice(fileMap, "apis")
		generateAPIDef(apis, call)
	}
}

func generateAPIDef(apis []interface{}, call *request.Request) {
	var definition map[string]interface{}
	for i := range apis {
		definition = map[string]interface{}{
			"api_definition": apis[i].(map[string]interface{})["api_definition"],
		}
		postAPI(definition, "/api/apis", call)
	}
}

func postAPI(definition map[string]interface{}, path string, call *request.Request) {
	api, err := json.Marshal(definition)
	utils.HandleError(err, false)
	req, err := call.FullRequest("POST", path, api)
	resp, err := call.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	} else {
		request.OutputResponse(resp)
	}
}
