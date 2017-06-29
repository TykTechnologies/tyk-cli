package cmd

import (
	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/cmd/usage"
	"github.com/TykTechnologies/tyk-cli/commands/remote"
)

var push bool
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add remotes to configuration",
	Long:  `Use this command to add remotes to configuration and post new organisations to Tyk`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			remote.Add("example.conf.json", args, push)
			return
		}
		cmd.Usage()
	},
}

func init() {
	remoteCmd.AddCommand(addCmd)
	addCmd.Flags().BoolVarP(&push, "push", "p", false, "Push new organisation to the Tyk instance")
	usage.Add(addCmd)
}
