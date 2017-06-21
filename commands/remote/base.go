package remote

import (
	"fmt"
	"io"
	"log"

	"github.com/TykTechnologies/tyk-cli/commands/api"
	"github.com/TykTechnologies/tyk-cli/db"
)

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

func ListApis(w io.Writer, conf []interface{}, args []string) {
	alias := args[0]
	var orgID string
	if len(args) > 0 {
		for _, remote := range conf {
			remote := remote.(map[string]interface{})
			if remote["alias"] == alias {
				orgID = remote["org_id"].(string)
			}
		}
	}
	bdb, err := db.OpenDB("bolt.db", 0444, true)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	list, err := api.FindByOrgID(bdb, orgID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "Staged APIs listed under Org. ID %v:\n", orgID)
	for _, api := range list {
		fmt.Fprintf(w, "%v - %v\n", api.APIDefinition.APIID, api.APIDefinition.Name)
	}
}
