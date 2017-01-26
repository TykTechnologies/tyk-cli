package importPkg

import (
	"encoding/json"
	"fmt"
	request "github.com/TykTechnologies/tyk-cli/request"
	utils "github.com/TykTechnologies/tyk-cli/utils"
	"io/ioutil"
	"os"
)

// Apis is a public function for importing APIs
func Apis(args []string) {
	if len(args) == 4 {
		call := request.New(args[0], args[1], args[2])
		uri := fmt.Sprintf("%s:%s/api/apis", call.Domain, call.Port)
		inputFile := args[3]
		parseJSON(inputFile, uri, call)
	}
}

func parseJSON(inputFile string, uri string, call *request.Request) {
	var fileObject interface{}
	file, _ := ioutil.ReadFile(utils.HandleFilePath(inputFile))
	err := json.Unmarshal([]byte(file), &fileObject)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	fileMap := utils.InterfaceToMap(fileObject)
	apis := fileMap["apis"].([]interface{})
	for i := range apis {
		definition := map[string]interface{}{
			"api_definition": utils.InterfaceToMap(
				apis[i],
			)["api_definition"],
		}
		postAPI(definition, uri, call)
	}
}

func postAPI(definition map[string]interface{}, uri string, call *request.Request) {
	api, id := apiAndID(definition)
	req, err := call.FullRequest("POST", uri, api)
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
		id = fmt.Sprintf("%v", utils.InterfaceToMap(definition["api_definition"])["id"])
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
