package _import

import (
	"bytes"
	"encoding/json"
	"fmt"
	utils "github.com/TykTechnologies/tyk-cli/utils"
	"io/ioutil"
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
		file, err := ioutil.ReadFile(utils.HandleFilePath("~/Documents/tykVagrant/example/apis37.json"))
		if err != nil {
			fmt.Printf("File error: %v\n", err)
			os.Exit(1)
		}
		payload, err := json.Marshal(file)
		fmt.Printf("PAYLOAD\n")
		fmt.Printf("PAYLOAD%v\n", bytes.NewBuffer(payload))
		req, err := utils.HttpRequest("POST", url, authorisation, payload)
		resp, err := client.Do(req)
		if err != nil {
			return
		} else {
			fmt.Printf("%v", resp)
		}
	}
}
