package _import

import (
	"bytes"
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
		url := fmt.Sprintf("%s:%s/api/apis", call.Domain, call.Port)
		file, err := ioutil.ReadFile(utils.HandleFilePath("~/Documents/tykVagrant/example/apis37.json"))
		if err != nil {
			fmt.Printf("File error: %v\n", err)
			os.Exit(1)
		}
		payload, err := json.Marshal(file)
		fmt.Printf("PAYLOAD\n")
		fmt.Printf("PAYLOAD%v\n", bytes.NewBuffer(payload))
		req, err := call.FullRequest("POST", url, payload)
		resp, err := call.Client.Do(req)
		if err != nil {
			return
		} else {
			fmt.Printf("%v", resp)
		}
	}
}
