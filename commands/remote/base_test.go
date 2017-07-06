package remote

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/TykTechnologies/tyk-cli/utils"
)

func TestAdd(t *testing.T) {
	fileName := "./test.conf.json"
	newAlias := "cats"
	err := wipeRemotes(fileName)
	if err != nil {
		t.Fatal(err)
	}
	err = Add(fileName, []string{newAlias, "http://tyk-docker.com:3000", "Dashboard"}, false)
	if err != nil {
		t.Fatal(err)
	}
	conf := utils.ParseJSONFile(fileName)
	remotes := conf["remotes"].([]interface{})
	var found map[string]interface{}
	for _, i := range remotes {
		remote := i.(map[string]interface{})
		if i.(map[string]interface{})["alias"] == newAlias {
			found = remote
		}
	}
	if len(found) == 0 {
		t.Fatal("Error: remote not created")
	}
}

func TestRemove(t *testing.T) {
	fileName := "./test.conf.json"
	newAlias := "cats"
	err := wipeRemotes(fileName)
	if err != nil {
		t.Fatal(err)
	}
	err = Add(fileName, []string{newAlias, "http://tyk-docker.com:3000", "Dashboard"}, false)
	if err != nil {
		t.Fatal(err)
	}
	err = Remove(fileName, []string{newAlias}, false)
	if err != nil {
		t.Fatal(err)
	}
	conf := utils.ParseJSONFile(fileName)
	remotes := conf["remotes"].([]interface{})
	var found map[string]interface{}
	for _, i := range remotes {
		remote := i.(map[string]interface{})
		if i.(map[string]interface{})["alias"] == newAlias {
			found = remote
		}
	}
	if len(found) != 0 {
		t.Fatal("Error: remote not removed")
	}
}

var remotes []interface{} = []interface{}{
	map[string]interface{}{
		"alias":             "default",
		"url":               "http://localhost:3000",
		"type":              "Dashboard",
		"organisation_name": "Default Org",
		"org_id":            "12345678901234567890",
		"auth_token":        "1234567890abcdef",
	},
	map[string]interface{}{
		"alias":             "catChannel",
		"url":               "http://localhost:8080",
		"type":              "Gateway",
		"organisation_name": "cat Org",
		"org_id":            "12345678901234567891",
		"auth_token":        "1234567890abcdef",
	},
}

func TestList(t *testing.T) {
	var buf bytes.Buffer
	List(&buf, remotes, false)
	result := buf.String()
	expectedResult := "default\ncatChannel\n"
	if result != expectedResult {
		t.Fatalf("Error - expected %s, got %s", expectedResult, result)
	}
}

func TestListVerbose(t *testing.T) {
	var buf bytes.Buffer
	List(&buf, remotes, true)
	result := buf.String()
	expectedResult := "Dashboard   default         http://localhost:3000\nGateway     catChannel      http://localhost:8080\n"
	if result != expectedResult {
		t.Fatalf("Error - expected:\n%s, got:\n%s", expectedResult, result)
	}
}

/*
 TODO - Failing test
func TestListAPIs(t *testing.T) {
	alias := remotes[0].(map[string]interface{})["alias"].(string)
	orgID := remotes[0].(map[string]interface{})["org_id"].(string)
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		t.Fatal(err)
	}
	defer bdb.Close()
	testAPI := api.New("Persia")
	err = testAPI.Create(bdb)
	if err != nil {
		t.Fatal(err)
	}
	err = testAPI.Edit(bdb, map[string]interface{}{"org_id": orgID})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	ListAPIs(&buf, remotes, []string{alias})
	result := buf.String()
	expectedResult := fmt.Sprintf("Staged APIs listed under Org. ID %v:\n%v - %v", orgID, testAPI.Id(), testAPI.Name())
	if result != expectedResult {
		t.Fatalf("Error - expected:\n%s, got:\n%s", expectedResult, result)
	}
	err = api.DeleteAll(bdb)
	if err != nil {
		t.Fatal(err)
	}
}
*/

func wipeRemotes(fileName string) error {
	conf := utils.ParseJSONFile(fileName)
	conf["remotes"] = []interface{}{}
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
