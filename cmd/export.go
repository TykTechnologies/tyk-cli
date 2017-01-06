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
	export "github.com/TykTechnologies/tyk-cli/commands/export"
	"github.com/spf13/cobra"
)

var domain, port, output string

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export API definitions",
	Long:  `This module lets you export API definitions and output to a JSON file`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		args = []string{key, domain, port, output}
		export.Apis(args)
		fmt.Println("export called")
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	exportCmd.Flags().StringVarP(&key, "key", "k", "", "Secret Key for the Dashboard API")
	exportCmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain name for your Dashboard")
	exportCmd.Flags().StringVarP(&port, "port", "p", "", "Port number for your Dashboard")
	exportCmd.Flags().StringVarP(&output, "output", "o", "", "Output file name for your JSON string")

}
