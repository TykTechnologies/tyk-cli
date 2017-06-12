package cmd

import (
	"os"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/commands/api"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Validate API definitions",
	Long:  `This is a subcommand of the 'api' command and can be used to test the validaity of an API.`,
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		api.Validate(id)
	},
}

func contains(args []string, item string) bool {
	for _, arg := range args {
		if arg == item {
			return true
		}
	}
	return false
}

func init() {
	apiCmd.AddCommand(testCmd)
}

func testUsage(cmd *cobra.Command) {
	cobra.AddTemplateFuncs(template.FuncMap{
		"add": func(i int, j int) int {
			return i + j
		},
		"parent": func() string {
			return os.Args[1]
		},
	})
	cmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  tyk-cli {{ parent }} [ID] test {{end}}{{if gt .Aliases 0}}
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
