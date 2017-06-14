package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/cmd/usage"
	"github.com/TykTechnologies/tyk-cli/commands/api"
	"github.com/TykTechnologies/tyk-cli/db"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new APIs.",
	Long:  `Create new APIs.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch len(args) {
		case 2:
			if args[0] == "api" {
				bdb, err := db.OpenDB("bolt.db", 0600, false)
				if err != nil {
					log.Fatal(err)
				}
				defer bdb.Close()
				name := args[1]
				newAPI := api.New(name)
				newAPI.Create(bdb)
				fmt.Printf("%v %v created ID %v\n", newAPI.Group(), newAPI.Name(), newAPI.Id())
			}
		default:
			usage.CreateAPI(cmd)
			cmd.Usage()
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	usage.CreateAPI(createCmd)
}
