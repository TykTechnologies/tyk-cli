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

	"github.com/spf13/cobra"
)

// bundleCmd represents the bundle command
var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Manage plugin bundles",
	Long: `This module lets you manage Tyk plugin bundles.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("bundle called")
	},
}

func init() {
	RootCmd.AddCommand(bundleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bundleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bundleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
