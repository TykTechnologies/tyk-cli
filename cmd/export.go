package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/commands/exportpkg"
)

var domain, port, output, id string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export API definitions",
	Long:  `This module lets you export API definitions and output to a JSON file`,
	Run: func(cmd *cobra.Command, args []string) {
		args = []string{id, key, domain, port, output}
		exportType := map[string]func([]string){
			"api": exportpkg.APIs,
		}
		exportType[cmd.Parent().Name()](args)
		fmt.Println("export called")
	},
}

func init() {
	apiCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&id, "id", "i", "", "API ID")
	exportCmd.Flags().StringVarP(&key, "key", "k", "", "Secret Key for the Dashboard API")
	exportCmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain name for your Dashboard")
	exportCmd.Flags().StringVarP(&port, "port", "p", "", "Port number for your Dashboard")
	exportCmd.Flags().StringVarP(&output, "output", "o", "", "Output file name for your JSON string")

}
