package _import

import (
	"fmt"
	request "github.com/TykTechnologies/tyk-cli/request"
	utils "github.com/TykTechnologies/tyk-cli/utils"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
)

type APIDefinition struct {
	Id  int
	Key string
}

type APIs struct {
	Collection []APIDefinition
}

func Apis(args []string) {
	if len(args) == 4 {
		call := request.New(args[0], args[1], args[2])
		url := fmt.Sprintf("%s:%s/api/apis", call.Domain, call.Port)
		file, err := ioutil.ReadFile(utils.HandleFilePath("~/Documents/tykVagrant/example/apis37.json"))
		if err != nil {
			fmt.Printf("File error: %v\n", err)
			os.Exit(1)
		}
		apis := gjson.GetBytes(file, "apis")
		for i := range apis.Array() {
			api := apis.Array()[i]
			payload := []byte(api.Raw)
			req, err := call.FullRequest("POST", url, payload)
			_, err = call.Client.Do(req)
			if err != nil {
				return
			} else {
				fmt.Printf(`
{
  "Status": "OK",
  "Message": "API created",
  "Meta": "%v"
}"
`, api.Get("api_definition.id"))
			}

		}
	}
}
