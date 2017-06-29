package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/cmd/usage"
	"github.com/TykTechnologies/tyk-cli/commands/remote"
	"github.com/TykTechnologies/tyk-cli/utils"
)

var apisCmd = &cobra.Command{
	Use:   "apis",
	Short: "List APIs within an organisation",
	Long:  "Use this command to list APIs within an organisation.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			conf := utils.ParseJSONFile("example.conf.json")["remotes"].([]interface{})
			remote.ListApis(os.Stdout, conf, args)
			return
		}
		cmd.Usage()
	},
}

func init() {
	remoteCmd.AddCommand(apisCmd)
	usage.APIs(apisCmd)
}
