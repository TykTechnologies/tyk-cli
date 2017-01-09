// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

	_import "github.com/TykTechnologies/tyk-cli/commands/_import"
	"github.com/spf13/cobra"
)

var input string

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import API definitions",
	Long:  `This module lets you import API definition from a JSON file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		args = []string{key, domain, port, output}
		_import.Apis(args)
		fmt.Println("import called")
	},
}

func init() {
	RootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importCmd.Flags().StringVarP(&key, "key", "k", "", "Secret Key for the Dashboard API")
	importCmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain name for your Dashboard")
	importCmd.Flags().StringVarP(&port, "port", "p", "", "Port number for your Dashboard")
	importCmd.Flags().StringVarP(&input, "input", "i", "", "Input file name for your JSON string")

}
