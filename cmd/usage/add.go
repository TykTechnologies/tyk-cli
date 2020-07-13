package usage

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/utils"
)

func Add(cmd *cobra.Command) {
	cmd.ResetCommands()
	usageFunc(cmd, addTemplate)
	if utils.Contains(os.Args, cmd.Name()) && utils.Contains(os.Args, "help") {
		cmd.Help()
		//Added to prevent duplicate help messages
		os.Exit(-1)
	}
}

var addTemplate string = `Usage:{{if .Runnable}}
  tyk-cli remote add {alias} {url} [{type}] {{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{if .HasAvailableSubCommands}}

Available Subcommands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name (add .NamePadding 5) }} {{.Short}}{{end}}{{end}}
  {{ end }}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
