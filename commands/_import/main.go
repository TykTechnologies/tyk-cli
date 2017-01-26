package _import

import (
	"encoding/json"
	"fmt"
	request "github.com/TykTechnologies/tyk-cli/request"
	utils "github.com/TykTechnologies/tyk-cli/utils"
	"io/ioutil"
	"os"
)

func Apis(args []string) {
	if len(args) == 4 {
		call := request.New(args[0], args[1], args[2])
		uri := fmt.Sprintf("%s:%s/api/apis", call.Domain, call.Port)
		input_file := args[3]
		parseJSON(input_file, uri, call)
	}
}

func parseJSON(input_file string, uri string, call *request.Request) {
	var f interface{}
	file, _ := ioutil.ReadFile(utils.HandleFilePath(input_file))
	err := json.Unmarshal([]byte(file), &f)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	m := utils.InterfaceToMap(f)
	apis := m["apis"].([]interface{})
	for i := range apis {
		definition := map[string]interface{}{
			"api_definition": utils.InterfaceToMap(apis[i])["api_definition"],
		}
		api, err := json.Marshal(definition)
		if err != nil {
			fmt.Println(err)
		} else {
			id := fmt.Sprintf("%v", utils.InterfaceToMap(definition["api_definition"])["id"])
			postAPI(api, id, uri, call)
		}
	}
}

func postAPI(api []byte, id string, uri string, call *request.Request) {
	req, err := call.FullRequest("POST", uri, api)
	_, err = call.Client.Do(req)
	if err != nil {
		return
	} else {
		apiCreatedMessage(id)
	}
}

func apiCreatedMessage(id string) {
	fmt.Printf(`{
  "Status": "OK",
  "Message": "API created",
  "Meta": "%s"
},
`, id)
}
