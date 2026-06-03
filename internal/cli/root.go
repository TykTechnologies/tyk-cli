package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tyktech/tyk-cli/internal/config"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// GlobalFlags holds global CLI flags
type GlobalFlags struct {
	DashURL   string
	AuthToken string
	OrgID     string
	JSON      bool
}

// reqproof:req REQ-API-001
func NewRootCommand(version, commit, buildTime string) *cobra.Command {
	var globalFlags GlobalFlags
	
	rootCmd := &cobra.Command{
		Use:          "tyk",
		Short:        "Tyk CLI - Manage Tyk OAS-native APIs",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Tyk CLI is a command-line interface for managing Tyk OAS-native APIs.
It provides commands to create, update, delete, and manage API versions
with support for OpenAPI 3.0 specifications.`,
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip configuration loading for setup and info commands
			skipCommands := []string{"version", "help", "init", "config"}
			for _, skipCmd := range skipCommands {
				if cmd.Name() == skipCmd || 
				   (cmd.Parent() != nil && cmd.Parent().Name() == skipCmd) ||
				   (cmd.Parent() != nil && cmd.Parent().Parent() != nil && cmd.Parent().Parent().Name() == skipCmd) {
					return nil
				}
			}
			
			return initConfig(cmd, &globalFlags)
		},
	}

	// Add version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`tyk version %s
  commit: %s
  built:  %s
`, version, commit, buildTime))

	// Add persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&globalFlags.DashURL, "dash-url", "", 
		"Tyk Dashboard URL (TYK_DASH_URL)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.AuthToken, "auth-token", "", 
		"Dashboard API auth token (TYK_AUTH_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.OrgID, "org-id", "", 
		"Organization ID (TYK_ORG_ID)")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.JSON, "json", false, 
		"Output in JSON format")

	// Add subcommands
	rootCmd.AddCommand(NewInitCommand())
	rootCmd.AddCommand(NewAPICommand())
	rootCmd.AddCommand(NewPolicyCommand())
	rootCmd.AddCommand(NewConfigCommand())

	return rootCmd
}

// reqproof:req REQ-CFG-001
func initConfig(cmd *cobra.Command, flags *GlobalFlags) error {
	// Create config manager
	configManager := config.NewManager()
	
	// Load config from environment and files
	if err := configManager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override with command line flags
	configManager.SetFromFlags(flags.DashURL, flags.AuthToken, flags.OrgID)

	// Validate configuration
	config := configManager.GetConfig()
	if err := config.Validate(); err != nil {
		return err
	}

	// Get effective config for API operations (resolves environment values)
	effectiveConfig := configManager.GetEffectiveConfig()

	// Store in command context
	cmd.SetContext(withConfig(cmd.Context(), effectiveConfig))
	cmd.SetContext(withOutputFormat(cmd.Context(), getOutputFormat(flags.JSON)))
	
	return nil
}

// reqproof:req REQ-CFG-001
func getOutputFormat(jsonFlag bool) types.OutputFormat {
	if jsonFlag {
		return types.OutputJSON
	}
	return types.OutputHuman
}

// reqproof:req REQ-CFG-001
func SetupViper() {
	viper.SetEnvPrefix("TYK")
	viper.AutomaticEnv()
}