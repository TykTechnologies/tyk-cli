package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/commands/bundle"
)

var buildOutput, key string
var skipSigning bool

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a plugin bundle",
	Long:  `This command will create a bundle, it will take a manifest file and its specified files as input.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := bundle.Bundle("build", buildOutput, key, &skipSigning)
		if err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	bundleCmd.AddCommand(buildCmd)

	buildCmd.PersistentFlags().StringVarP(&buildOutput, "output", "o", "", "Output file")
	buildCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "Key for bundle signature")
	buildCmd.PersistentFlags().BoolVarP(&skipSigning, "skip-signing", "y", false, "Skip bundle signing")
}
