// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	bundle "github.com/TykTechnologies/tyk-cli/commands/bundle"
	"github.com/spf13/cobra"
)

var buildOutput, key string
var skipSigning bool

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a plugin bundle",
	Long:  `This command will create a bundle, it will take a manifest file and its specified files as input.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		err = bundle.Bundle("build", buildOutput, key, &skipSigning)
		if err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	bundleCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("output", "", "Output file")
	buildCmd.PersistentFlags().StringVarP(&buildOutput, "output", "o", "", "Output file")
	buildCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "Key for bundle signature")
	buildCmd.PersistentFlags().BoolP("skip-signing", "y", false, "Skip bundle signing")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
