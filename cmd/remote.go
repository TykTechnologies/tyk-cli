package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/cmd/usage"
	"github.com/TykTechnologies/tyk-cli/commands/remote"
	"github.com/TykTechnologies/tyk-cli/utils"
)

var verbose bool
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Select a remote",
	Long:  "Select a remote",
	Run: func(cmd *cobra.Command, args []string) {
		switch len(args) {
		case 0:
			conf := utils.ParseJSONFile("example.conf.json")["remotes"].([]interface{})
			remote.List(os.Stdout, conf, verbose)
		case 1:
			fmt.Printf("unknown remote subcommand: %s\n", args[0])
			cmd.Usage()
		default:
			if utils.Contains([]string{"add", "remove", "rm"}, args[0]) {
				remSubCmds(cmd, args)
				return
			}
			aliasSubCmds(cmd, args)
		}
	},
}

func remSubCmds(cmd *cobra.Command, args []string) {
	subCmd := args[0]
	switch subCmd {
	case "add":
		addCmd.Run(addCmd, args[1:])
	case "rm":
		rmCmd.Run(rmCmd, args[1:])
	case "remove":
		rmCmd.Run(rmCmd, args[1:])
	default:
		fmt.Printf("unknown remote subcommand: %s\n", args[0])
		cmd.Usage()
	}
}

func aliasSubCmds(cmd *cobra.Command, args []string) {
	alias := args[0]
	subCmd := args[1]
	switch subCmd {
	case "apis":
		apis := append([]string{alias}, args[2:]...)
		apisCmd.Run(apisCmd, apis)
	default:
		fmt.Printf("unknown remote subcommand: %s\n", args[0])
		cmd.Usage()
	}
}

func init() {
	RootCmd.AddCommand(remoteCmd)
	remoteCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "List available remotes and URLs")
	usage.Remote(remoteCmd)
}
