package cmd

import (
	"github.com/spf13/cobra"
)

var apiId string

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Manage API definitions",
	Long:  `This module lets you manage API definitions using the Dashboard API.`,
	Run: func(cmd *cobra.Command, args []string) {
		args = []string{apiId}
		cmd.Usage()
	},
}

func init() {
	RootCmd.AddCommand(apiCmd)

	apiCmd.Flags().StringVarP(&apiId, "apiId", "", "", "API ID")
}
