package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/cmd/usage"
	"github.com/TykTechnologies/tyk-cli/commands/api"
	"github.com/TykTechnologies/tyk-cli/db"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit API definitions",
	Long:  `Use this command to update individual API definitions by inputting a JSON string`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 2 {
			id := args[0]
			var params map[string]interface{}
			err := json.Unmarshal([]byte(args[1]), &params)
			if err != nil {
				log.Fatal(err)
			}

			bdb, err := db.OpenDB("bolt.db", 0600, false)
			if err != nil {
				log.Fatal(err)
			}
			defer bdb.Close()
			e, err := api.Find(bdb, id)
			if err != nil {
				log.Fatal(err)
			}
			e.Edit(bdb, params)
			fmt.Printf("Edited API ID %v with the following attributes:\n%v\n", id, args[1])
			return
		}
		cmd.Usage()
	},
}

func init() {
	apiCmd.AddCommand(editCmd)
	usage.Edit(editCmd)
}
