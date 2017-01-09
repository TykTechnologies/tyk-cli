package _import

import (
	"fmt"
	request "github.com/TykTechnologies/tyk-cli/request"
	utils "github.com/TykTechnologies/tyk-cli/utils"
	"github.com/tidwall/gjson"
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
	file, err := ioutil.ReadFile(utils.HandleFilePath(input_file))
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	apis := gjson.GetBytes(file, "apis")
	for i := range apis.Array() {
		api := apis.Array()[i]
		postAPI(api, uri, call)
	}
}

func postAPI(api gjson.Result, uri string, call *request.Request) {
	payload := []byte(api.Raw)
	req, err := call.FullRequest("POST", uri, payload)
	_, err = call.Client.Do(req)
	if err != nil {
		return
	} else {
		apiCreatedMessage(api.Get("api_definition.id"))
	}
}

func apiCreatedMessage(id gjson.Result) {
	fmt.Printf(`{
  "Status": "OK",
  "Message": "API created",
  "Meta": "%v"
},
`, id)
}
