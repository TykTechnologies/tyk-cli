package cmd

import (
	"github.com/TykTechnologies/tyk-cli/cmd/usage"
	"github.com/TykTechnologies/tyk-cli/commands/remote"
	"github.com/spf13/cobra"
)

var force bool
var rmCmd = &cobra.Command{
	Use:     "rm",
	Aliases: []string{"remove"},
	Short:   "Remove remotes from configuration",
	Long:    `Use this command to remove remotes from the configuration file and remove organisations from a Tyk instance`,
	Run: func(cmd *cobra.Command, args []string) {
		remote.Remove("example.conf.json", args, force)
	},
}

func init() {
	remoteCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolVarP(&force, "force", "f", false, "Remove organisation from a Tyk instance")
	usage.Remove(rmCmd)
}
