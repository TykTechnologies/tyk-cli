package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/config"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// NewConfigCommand creates the 'tyk config' command for unified environment management
func NewConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration environments",
		Long: `Unified environment and configuration management.

In the unified approach, environments ARE the configuration system. 
Each environment contains dashboard URL, auth token, and org ID.

Examples:
  tyk config list                    # List all environments
  tyk config use staging             # Switch to staging environment  
  tyk config current                 # Show current environment
  tyk config add dev --dashboard-url http://localhost:3000 --auth-token token --org-id org
  tyk config set dashboard-url https://api.tyk.io  # Update current environment`,
	}

	configCmd.AddCommand(NewConfigListCommand())
	configCmd.AddCommand(NewConfigUseCommand())
	configCmd.AddCommand(NewConfigCurrentCommand())
	configCmd.AddCommand(NewConfigAddCommand())
	configCmd.AddCommand(NewConfigSetCommand())
	configCmd.AddCommand(NewConfigRemoveCommand())

	return configCmd
}

// NewConfigListCommand creates the 'tyk config list' command
func NewConfigListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured environments",
		Long:  "Display all configured environments and show which one is active",
		RunE:  runConfigList,
	}

	return cmd
}

// NewConfigUseCommand creates the 'tyk config use' command  
func NewConfigUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use [environment-name]",
		Short: "Switch to a different environment",
		Long:  "Change the active environment. If no environment name is provided, you'll be prompted to select from available environments.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runConfigUse,
	}

	return cmd
}

// NewConfigCurrentCommand creates the 'tyk config current' command
func NewConfigCurrentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show current active environment",
		Long:  "Display the currently active environment and its configuration",
		RunE:  runConfigCurrent,
	}

	return cmd
}

// NewConfigAddCommand creates the 'tyk config add' command
func NewConfigAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <environment-name>",
		Short: "Add a new environment",
		Long: `Add a new environment configuration.

Examples:
  tyk config add development --dashboard-url http://localhost:3000 --auth-token token --org-id org
  tyk config add production --dashboard-url https://prod-dashboard.com --auth-token prod-token --org-id prod-org --set-default`,
		Args: cobra.ExactArgs(1),
		RunE: runConfigAdd,
	}

	cmd.Flags().String("dashboard-url", "", "Tyk Dashboard URL")
	cmd.Flags().String("auth-token", "", "Dashboard API auth token")
	cmd.Flags().String("org-id", "", "Organization ID")
	cmd.Flags().Bool("set-default", false, "Set this environment as the default")

	_ = cmd.MarkFlagRequired("dashboard-url")
	_ = cmd.MarkFlagRequired("auth-token")
	_ = cmd.MarkFlagRequired("org-id")

	return cmd
}

// NewConfigSetCommand creates the 'tyk config set' command
func NewConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update current environment configuration",
		Long: `Update configuration values for the current active environment.

Examples:
  tyk config set dashboard-url https://new-dashboard.com
  tyk config set auth-token new-token
  tyk config set org-id new-org-id
  
  # Set multiple values at once
  tyk config set dashboard-url https://api.tyk.io auth-token token org-id org`,
		RunE: runConfigSet,
	}

	cmd.Flags().String("dashboard-url", "", "Update dashboard URL")
	cmd.Flags().String("auth-token", "", "Update auth token")  
	cmd.Flags().String("org-id", "", "Update organization ID")

	return cmd
}

// NewConfigRemoveCommand creates the 'tyk config remove' command
func NewConfigRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <environment-name>",
		Short: "Remove an environment",
		Long:  "Remove an environment from the configuration",
		Args:  cobra.ExactArgs(1),
		RunE:  runConfigRemove,
	}

	return cmd
}

func runConfigList(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	environments := manager.ListEnvironments()

	if len(environments) == 0 {
		yellow := color.New(color.FgYellow)
		yellow.Println("No environments configured.")
		fmt.Println("Use 'tyk config add <name> ...' to add an environment.")
		return nil
	}

	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	
	blue.Printf("Default environment: ")
	green.Printf("%s\n\n", cfg.DefaultEnvironment)
	
	blue.Println("Environments:")

	// Sort environment names for consistent display
	var envNames []string
	for name := range environments {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)

	for _, name := range envNames {
		env := environments[name]
		if name == cfg.DefaultEnvironment {
			green.Printf("● %s (active):\n", name)
		} else {
			fmt.Printf("  %s:\n", name)
		}
		cyan.Printf("    dashboard_url = %s\n", env.DashboardURL)
		cyan.Printf("    auth_token    = %s\n", maskToken(env.AuthToken))
		cyan.Printf("    org_id        = %s\n", env.OrgID)
		fmt.Println()
	}

	return nil
}

func runConfigUse(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	environments := manager.ListEnvironments()

	if len(environments) == 0 {
		return fmt.Errorf("no environments configured. Use 'tyk config add' to add an environment")
	}

	var envName string
	
	// If environment name was provided as argument, use it
	if len(args) > 0 {
		envName = args[0]
	} else {
		// Interactive selection
		var err error
		envName, err = selectEnvironmentInteractively(environments, cfg.DefaultEnvironment)
		if err != nil {
			return err
		}
		if envName == "" {
			return fmt.Errorf("no environment selected")
		}
	}

	// Check if environment exists
	if _, err := manager.GetEnvironment(envName); err != nil {
		return err
	}

	// Set as default
	if err := manager.SetDefaultEnvironment(envName); err != nil {
		return err
	}

	// Save to file
	if err := saveConfigToFile(manager); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✓ Switched to environment '%s'.\n", envName)
	return nil
}

func runConfigCurrent(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	
	if cfg.DefaultEnvironment == "" {
		yellow := color.New(color.FgYellow)
		yellow.Println("No default environment set.")
		return nil
	}

	activeEnv, err := cfg.GetActiveEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get active environment: %w", err)
	}

	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)

	blue.Println("Current environment:")
	green.Printf("● %s (active)\n", activeEnv.Name)
	cyan.Printf("  dashboard_url = %s\n", activeEnv.DashboardURL)
	cyan.Printf("  auth_token    = %s\n", maskToken(activeEnv.AuthToken))
	cyan.Printf("  org_id        = %s\n", activeEnv.OrgID)

	return nil
}

func runConfigAdd(cmd *cobra.Command, args []string) error {
	envName := args[0]
	dashboardURL, _ := cmd.Flags().GetString("dashboard-url")
	authToken, _ := cmd.Flags().GetString("auth-token")
	orgID, _ := cmd.Flags().GetString("org-id")
	setDefault, _ := cmd.Flags().GetBool("set-default")

	// Create the environment
	env := &types.Environment{
		Name:         envName,
		DashboardURL: dashboardURL,
		AuthToken:    authToken,
		OrgID:        orgID,
	}

	// Validate the environment
	if err := env.Validate(); err != nil {
		return err
	}

	// Load existing configuration
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if environment already exists
	if _, err := manager.GetEnvironment(envName); err == nil {
		return fmt.Errorf("environment '%s' already exists. Use 'tyk config set' to update it", envName)
	}

	// Save the environment
	if err := manager.SaveEnvironment(env, setDefault); err != nil {
		return fmt.Errorf("failed to save environment: %w", err)
	}

	// Save to file
	if err := saveConfigToFile(manager); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✓ Environment '%s' added successfully.\n", envName)
	if setDefault || manager.GetConfig().DefaultEnvironment == envName {
		green.Printf("✓ Environment '%s' set as default.\n", envName)
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	dashboardURL, _ := cmd.Flags().GetString("dashboard-url")
	authToken, _ := cmd.Flags().GetString("auth-token")
	orgID, _ := cmd.Flags().GetString("org-id")

	if dashboardURL == "" && authToken == "" && orgID == "" {
		return fmt.Errorf("at least one configuration value must be provided")
	}

	// Load configuration
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	
	// Get active environment
	if cfg.DefaultEnvironment == "" {
		return fmt.Errorf("no active environment. Use 'tyk config add' to create one")
	}

	activeEnv, err := manager.GetEnvironment(cfg.DefaultEnvironment)
	if err != nil {
		return err
	}

	// Update the active environment
	if dashboardURL != "" {
		activeEnv.DashboardURL = dashboardURL
	}
	if authToken != "" {
		activeEnv.AuthToken = authToken
	}
	if orgID != "" {
		activeEnv.OrgID = orgID
	}

	// Validate updated environment
	if err := activeEnv.Validate(); err != nil {
		return err
	}

	// Save to file
	if err := saveConfigToFile(manager); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✓ Environment '%s' updated successfully.\n", activeEnv.Name)
	
	// Show what was updated
	if dashboardURL != "" {
		fmt.Printf("  dashboard_url = %s\n", dashboardURL)
	}
	if authToken != "" {
		fmt.Printf("  auth_token    = %s\n", maskToken(authToken))
	}
	if orgID != "" {
		fmt.Printf("  org_id        = %s\n", orgID)
	}

	return nil
}

func runConfigRemove(cmd *cobra.Command, args []string) error {
	envName := args[0]

	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	// Check if environment exists
	if cfg.Environments[envName] == nil {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	// Don't allow removing the last environment
	if len(cfg.Environments) == 1 {
		return fmt.Errorf("cannot remove the last environment")
	}

	// Remove environment
	delete(cfg.Environments, envName)

	// If this was the default, pick another one
	if cfg.DefaultEnvironment == envName {
		for name := range cfg.Environments {
			cfg.DefaultEnvironment = name
			break
		}
		yellow := color.New(color.FgYellow)
		yellow.Printf("⚠ Default environment changed to '%s'.\n", cfg.DefaultEnvironment)
	}

	// Save to file
	if err := saveConfigToFile(manager); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✓ Environment '%s' removed successfully.\n", envName)
	return nil
}

func saveConfigToFile(manager *config.Manager) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	cfg := manager.GetConfig()
	content := generateTOMLConfigUnified(cfg)

	configFile := filepath.Join(configDir, "cli.toml")
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userConfigDir, "tyk"), nil
}

func generateTOMLConfigUnified(cfg *types.Config) string {
	content := "# Tyk CLI Configuration\n"
	content += "# This file stores named environments for the Tyk CLI\n"
	content += "# In the unified approach, environments ARE the configuration system\n\n"
	
	// Set default environment
	if cfg.DefaultEnvironment != "" {
		content += fmt.Sprintf("default_environment = \"%s\"\n\n", cfg.DefaultEnvironment)
	}
	
	// Add all environments
	if len(cfg.Environments) > 0 {
		for name, env := range cfg.Environments {
			content += fmt.Sprintf("[environments.%s]\n", name)
			content += fmt.Sprintf("name = \"%s\"\n", env.Name)
			content += fmt.Sprintf("dashboard_url = \"%s\"\n", env.DashboardURL)
			content += fmt.Sprintf("auth_token = \"%s\"\n", env.AuthToken)
			content += fmt.Sprintf("org_id = \"%s\"\n", env.OrgID)
			content += "\n"
		}
	}
	
	return content
}

func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

func selectEnvironmentInteractively(environments map[string]*types.Environment, currentDefault string) (string, error) {
	// Create sorted list of environment names for consistent display
	var envNames []string
	for name := range environments {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)

	// Create display options with colors and current default indicator
	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	
	var options []string
	for _, name := range envNames {
		env := environments[name]
		var displayName string
		if name == currentDefault {
			displayName = green.Sprintf("● %s (current)", name)
		} else {
			displayName = fmt.Sprintf("  %s", name)
		}
		
		// Add environment details
		displayName += yellow.Sprintf(" - %s", env.DashboardURL)
		options = append(options, displayName)
	}

	// Create the interactive prompt
	prompt := &survey.Select{
		Message: blue.Sprint("Select environment to switch to:"),
		Options: options,
		Help:    "Use arrow keys to navigate, Enter to select, Ctrl+C to cancel",
	}

	var selectedIndex int
	if err := survey.AskOne(prompt, &selectedIndex); err != nil {
		return "", fmt.Errorf("environment selection cancelled: %w", err)
	}

	return envNames[selectedIndex], nil
}