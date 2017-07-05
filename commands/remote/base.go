package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/TykTechnologies/tyk-cli/commands/api"
	"github.com/TykTechnologies/tyk-cli/db"
	"github.com/TykTechnologies/tyk-cli/request"
	"github.com/TykTechnologies/tyk-cli/utils"
)

func Add(fileName string, args []string, push bool) error {
	conf := utils.ParseJSONFile(fileName)
	remotes := conf["remotes"].([]interface{})
	alias := args[0]
	remType := "Dashboard"
	uri := args[1]
	adminAuth := ifTrueReturn(push, conf["admin-auth"].(string))
	var orgID string
	var err error
	if push {
		orgID, err = createOrg(uri, adminAuth, alias)
	} else {
		orgID = "Default Org."
	}
	if err != nil {
		return err
	}
	if len(args) == 3 {
		remType = args[2]
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

func createOrg(uri, adminAuth, owner string) (string, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return "", err
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
		return "", err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/admin/organisations/", uri), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("admin-auth", adminAuth)
	resp, err := call.Client.Do(req)
	if err != nil {
		return "", err
	}
	var respBody map[string]interface{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(b, &respBody)
	if err != nil {
		return "", err
	}
	return respBody["Meta"].(string), nil
}

func ifTrueReturn(t bool, value string) string {
	if t {
		return value
	}
	return ""
}

func Remove(fileName string, args []string, force bool) error {
	conf := utils.ParseJSONFile(fileName)
	remotes := conf["remotes"].([]interface{})
	alias := args[0]
	for _, v := range remotes {
		if v.(map[string]interface{})["alias"].(string) == alias {
			if force {
				adminAuth := conf["admin-auth"].(string)
				orgID := v.(map[string]interface{})["org_id"].(string)
				uri := v.(map[string]interface{})["url"].(string)
				err := deleteOrg(uri, adminAuth, orgID)
				if err != nil {
					return err
				}
				fmt.Printf("Deleted organisation %v\n", alias)
			}
			remotes[len(remotes)-1], v = v, remotes[len(remotes)-1]
			remotes = remotes[:len(remotes)-1]
		}
	}
	conf["remotes"] = remotes
	newConf, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, newConf, 0644)
	if err != nil {
		return err
	}
	return nil
}

func deleteOrg(uri, adminAuth, orgID string) error {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return err
	}
	host := fmt.Sprintf("%s://%s", u.Scheme, u.Hostname())
	call := request.New(adminAuth, host, u.Port())
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/admin/organisations/%s", uri, orgID), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("admin-auth", adminAuth)
	resp, err := call.Client.Do(req)
	if err != nil {
		return err
	}
	var respBody map[string]interface{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &respBody)
	if err != nil {
		return err
	}
	return nil
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
