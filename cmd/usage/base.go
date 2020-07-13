package usage

import (
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

func addTemplatesFuncs() {
	cobra.AddTemplateFuncs(template.FuncMap{
		"add": func(i, j int) int {
			return i + j
		},
		"parent": func() string {
			return os.Args[1]
		},
	})
}

func usageFunc(cmd *cobra.Command, template string) {
	addTemplatesFuncs()
	cmd.SetUsageTemplate(template)
	cmd.SetHelpTemplate(cmd.Long + "\n\n" + template)
}
