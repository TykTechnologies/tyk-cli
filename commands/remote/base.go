package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/TykTechnologies/tyk-cli/commands/api"
	"github.com/TykTechnologies/tyk-cli/db"
	"github.com/TykTechnologies/tyk-cli/request"
	"github.com/TykTechnologies/tyk-cli/utils"
)

func Add(fileName string, args []string) error {
	conf := utils.ParseJSONFile(fileName)
	remotes := conf["remotes"].([]interface{})
	adminAuth := conf["admin-auth"].(string)
	alias := args[0]
	uri := args[1]
	remType := "Dashboard"
	if len(args) == 3 {
		remType = args[2]
	}
	orgID, err := createOrg(uri, adminAuth, alias)
	if err != nil {
		log.Fatal(err)
	}
	remote := map[string]interface{}{
		"alias":             alias,
		"type":              remType,
		"url":               uri,
		"organisation_name": alias,
		"org_id":            orgID,
		"auth_token":        "",
	}
	conf["remotes"] = append(remotes, remote)
	newConf, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, newConf, 0644)
	if err != nil {
		return err
	}
	fmt.Printf("Created organisation ID %v\n", orgID)
	return nil
}

func createOrg(uri, adminAuth, owner string) (interface{}, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}
	host := fmt.Sprintf("%v://%s", u.Scheme, u.Hostname())
	call := request.New(adminAuth, host, u.Port())
	newOrg := map[string]interface{}{
		"owner_name":    owner,
		"cname":         u.Hostname(),
		"cname_enabled": true,
	}
	payload, err := json.Marshal(newOrg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/admin/organisations/", uri), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("admin-auth", adminAuth)
	resp, err := call.Client.Do(req)
	if err != nil {
		return nil, err
	}
	var respBody map[string]interface{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &respBody)
	if err != nil {
		return nil, err
	}
	return respBody["Meta"], nil
}

func List(w io.Writer, conf []interface{}, verbose bool) {
	for _, remote := range conf {
		remote := remote.(map[string]interface{})
		if verbose {
			fmt.Fprintf(w, "%-10v  %-15v %v\n", remote["type"], remote["alias"], remote["url"])
		} else {
			fmt.Fprintf(w, "%v\n", remote["alias"])
		}
	}
}

func ListApis(w io.Writer, conf []interface{}, args []string) error {
	alias := args[0]
	var orgID string
	for _, remote := range conf {
		remote := remote.(map[string]interface{})
		if remote["alias"] == alias {
			orgID = remote["org_id"].(string)
		}
	}
	bdb, err := db.OpenDB("bolt.db", 0444, true)
	if err != nil {
		return err
	}
	defer bdb.Close()
	list, err := api.FindByOrgID(bdb, orgID)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Staged APIs listed under Org. ID %v:\n", orgID)
	for _, api := range list {
		fmt.Fprintf(w, "%v - %v\n", api.APIDefinition.APIID, api.APIDefinition.Name)
	}
	return nil
}
