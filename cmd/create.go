package cmd

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/commands/api"
	"github.com/TykTechnologies/tyk-cli/db"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new APIs.",
	Long:  `Create new APIs.`,
	Run: func(cmd *cobra.Command, args []string) {
		args = os.Args[2:]
		switch len(args) {
		case 1:
			subCmdUsage(cmd, args)
		case 2:
			if args[0] == "api" {
				bdb, err := db.OpenDB("bolt.db", 0600, false)
				if err != nil {
					log.Fatal(err)
				}
				defer bdb.Close()
				name := args[1]
				newAPI := api.New(name)
				newAPI.Create(bdb)
				fmt.Printf("%v %v created ID %v\n", newAPI.Group(), newAPI.Name(), newAPI.Id())
			}
		default:
			createAPIUsage(cmd)
		}
	},
}

func subCmdUsage(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "api":
		createAPIUsage(cmd)
	default:
		log.Println("Please implement me")
	}
}

func createAPIUsage(cmd *cobra.Command) {
	cobra.AddTemplateFuncs(template.FuncMap{
		"add": func(i int, j int) int {
			return i + j
		},
	})
	cmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{ .CommandPath }} api [name] {{end}}{{if gt .Aliases 0}}
Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}
Examples:
{{ .Example }}{{end}}{{if .HasAvailableSubCommands}}
Available Subcommands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}{{ if (eq .Name "test") }}
  [ID] {{rpad .Name .NamePadding }} {{.Short}}{{ else }}
  {{rpad .Name (add .NamePadding 5) }} {{.Short}}{{end}}{{end}}{{end}}
  {{ end }}{{if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
	cmd.Usage()
}

func init() {
	if contains(os.Args, "--help") || contains(os.Args, "-h") {
		createAPIUsage(createCmd)
	}
	RootCmd.AddCommand(createCmd)
}
