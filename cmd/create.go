package cmd

import (
	"os"
	"strings"
	"text/template"

	"github.com/TykTechnologies/tyk-cli/commands/api"
	"github.com/TykTechnologies/tyk-cli/db"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new APIs.",
	Long:  `Create new APIs.`,
	Run: func(cmd *cobra.Command, args []string) {
		args = os.Args[2:]
		switch len(args) {
		case 0:
			createAPIUsage(cmd)
		case 1:
			switch args[0] {
			case "api":
				createAPIUsage(cmd)
			}
		case 2:
			if args[0] == "api" {
				name := args[1]
				apiId := strings.Replace(uuid.NewV4().String(), "-", "", -1)
				API := api.New()
				API.Create(&db.Item{apiId, name})
			}
		}
	},
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
