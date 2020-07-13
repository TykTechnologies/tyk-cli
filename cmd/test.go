package cmd

import (
	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/cmd/usage"
	"github.com/TykTechnologies/tyk-cli/commands/api"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Validate API definitions",
	Long: `This is a subcommand of the 'api' command and can be used to test the validity
of an API.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			id := args[0]
			api.Validate(id)
			return
		}
		cmd.Usage()
	},
}

func init() {
	apiCmd.AddCommand(testCmd)
	usage.TestUsage(testCmd)
}
