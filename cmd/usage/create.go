package usage

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/TykTechnologies/tyk-cli/utils"
)

func CreateAPI(cmd *cobra.Command) {
	//cmd.ResetCommands()
	usageFunc(cmd, createAPITemplate)
	if utils.Contains(os.Args, cmd.Name()) {
		for _, i := range []string{"help", "-h", "--help"} {
			if utils.Contains(os.Args, i) {
				cmd.Help()
				//Added to prevent duplicate help messages
				os.Exit(-1)
			}
		}
	}
}

var createAPITemplate string = `Usage:{{if .Runnable}}
  {{ .CommandPath }} [command] [name] {{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{if .HasAvailableSubCommands}}

Available Subcommands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{ end }}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
