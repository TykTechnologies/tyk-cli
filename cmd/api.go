package cmd

import (
	"fmt"
	"os"

	"github.com/TykTechnologies/tyk-cli/cmd/usage"
	"github.com/TykTechnologies/tyk-cli/utils"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Manage API definitions",
	Long:  `This module lets you manage API definitions using the Dashboard API.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Errorf("need to specify api id or subcommand")
			cmd.Usage()
			return
		}
		apiID := args[0]
		if len(args) == 1 {
			cmd.ResetCommands()
			if !utils.Contains([]string{"help", "test"}, apiID) && !utils.Contains(os.Args, "create") {
				fmt.Printf("selected api %s, please add subcommand\n", apiID)
				cmd.Usage()
			}
			if apiID == "help" {
				cmd.Help()
			}
			return
		}
		subCmd := args[1]
		apiSubCmds(apiID, subCmd)
	},
}

func apiSubCmds(apiID, subCmd string) {
	switch subCmd {
	case "test":
		testCmd.Run(testCmd, []string{apiID})
	default:
		fmt.Printf("unknown api subcommand: %s\n", subCmd)
	}
}

func init() {
	if utils.Contains(os.Args, "create") && utils.Contains(os.Args, "help") {
		createCmd.AddCommand(apiCmd)
		usage.CreateAPI(createCmd)
		return
	}
	RootCmd.AddCommand(apiCmd)
	usage.API(apiCmd)
}
