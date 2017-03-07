package importpkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/TykTechnologies/tyk-cli/request"
	"github.com/TykTechnologies/tyk-cli/utils"
)

// Apis is a public function for importing APIs
func Apis(args []string) {
	if len(args) == 4 {
		call := request.New(args[0], args[1], args[2])
		inputFile := args[3]
		parseJSON(inputFile, "/api/apis", call)
	}
}

func parseJSON(inputFile string, path string, call *request.Request) {
	var fileObject interface{}
	file, _ := ioutil.ReadFile(utils.HandleFilePath(inputFile))
	err := json.Unmarshal([]byte(file), &fileObject)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	fileMap := fileObject.(map[string]interface{})
	apis := fileMap["apis"].([]interface{})
	for i := range apis {
		definition := map[string]interface{}{
			"api_definition": apis[i].(map[string]interface{})["api_definition"],
		}
		postAPI(definition, path, call)
	}
}

func postAPI(definition map[string]interface{}, path string, call *request.Request) {
	api, id := apiAndID(definition)
	req, err := call.FullRequest("POST", path, api)
	_, err = call.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	} else {
		apiCreatedMessage(id)
	}
}

func apiAndID(definition map[string]interface{}) (api []byte, id string) {
	api, err := json.Marshal(definition)
	if err != nil {
		fmt.Println(err)
	} else {
		id = fmt.Sprintf("%v", definition["api_definition"].(map[string]interface{})["id"])
	}
	return
}

func apiCreatedMessage(id string) {
	fmt.Printf(`{
  "Status": "OK",
  "Message": "API created",
  "Meta": "%s"
},
`, id)
}
