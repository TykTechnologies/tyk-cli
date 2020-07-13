package usage

import (
	"github.com/spf13/cobra"
)

func API(cmd *cobra.Command) {
	cmd.ResetCommands()
	usageFunc(cmd, apiTemplate)
}

var apiTemplate string = `Usage:{{if .Runnable}}
  {{ .CommandPath}} [ID] [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{if .HasAvailableSubCommands}}

Available Subcommands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}{{ if (eq .Name "test") }}
  [ID] {{rpad .Name (add .NamePadding 12) }} {{.Short}}{{ else if (eq .Name "edit") }}
  [ID] {{ .Name  }} '{"key": "value"}' {{.Short}}{{ else }}
  {{rpad .Name (add .NamePadding 17) }} {{.Short}}{{end}}{{end}}{{end}}{{ end }}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
