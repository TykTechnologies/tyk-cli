package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Manage API definitions",
	Long:  `This module lets you manage API definitions using the Dashboard API.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiUsage(cmd, true)
		args = os.Args[2:]
		if len(args) < 1 {
			fmt.Errorf("need to specify api id or subcommand")
			apiUsage(cmd, false)
			return
		}
		apiId := args[0]
		if len(args) == 1 {
			fmt.Printf("selected api %s, please add subcommand\n", apiId)
			apiUsage(cmd, true)
			return
		}
		subCmd := args[1]
		switch subCmd {
		case "test":
			testCmd.Run(testCmd, []string{apiId})
		default:
			fmt.Errorf("unknown api subcommand: %s", args[0])
		}
	},
}

func apiUsage(cmd *cobra.Command, isSubCmd bool) {
	if isSubCmd {
		cmd.ResetCommands()
	}
	cobra.AddTemplateFuncs(template.FuncMap{
		"add": func(i int, j int) int {
			return i + j
		},
	})
	cmd.AddCommand(testCmd)
	cmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{ .CommandPath}} [ID] [command]{{end}}{{if gt .Aliases 0}}
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
	if contains(os.Args, "test") && (contains(os.Args, "--help") || contains(os.Args, "-h")) {
		RootCmd.AddCommand(testCmd)
		testUsage(testCmd)
		os.Exit(-1)
	} else {
		RootCmd.AddCommand(apiCmd)
		apiCmd.AddCommand(testCmd)
	}
}
