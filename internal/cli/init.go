package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/client"
	"github.com/tyktech/tyk-cli/internal/config"
	"github.com/tyktech/tyk-cli/pkg/types"
)


// Implements: SYS-REQ-042
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
    Use:   "init",
    Short: "Interactive setup wizard for Tyk CLI",
    Long: `🚀 Interactive setup wizard to get you started quickly!

This wizard will help you:
- Configure your Tyk Dashboard connection
- Bootstrap a single environment (e.g., dev)
- Optionally test the connection
- Save everything for future use`,
		RunE: runInitWizard,
	}

	cmd.Flags().Bool("skip-test", false, "Skip connection testing")
	cmd.Flags().Bool("quick", false, "Quick setup (single environment)")

	return cmd
}

// Implements: SYS-REQ-042
func runInitWizard(cmd *cobra.Command, args []string) error {
    skipTest, _ := cmd.Flags().GetBool("skip-test")
    // quick flag retained for compatibility; the wizard now always bootstraps a single env
    _, _ = cmd.Flags().GetBool("quick")

	scanner := bufio.NewScanner(os.Stdin)

	printWelcome()
	
    // Always run single-environment setup
    return runQuickSetup(scanner, skipTest)
}

// Implements: SYS-REQ-042
func printWelcome() {
    fmt.Println("🚀 Welcome to Tyk CLI Setup Wizard!")
    fmt.Println("====================================")
    fmt.Println()
    fmt.Println("This wizard configures a single environment (e.g., dev) so you can get going fast.")
    fmt.Println("You can add more environments later with 'tyk config add'.")
    fmt.Println()
}

// Implements: SYS-REQ-042
func runQuickSetup(scanner *bufio.Scanner, skipTest bool) error {
    fmt.Println("⚡ Quick Setup Mode")
    fmt.Println("------------------")
    fmt.Println()

    env, err := gatherEnvironmentInfo(scanner, "dev", true)
    if err != nil {
        return err
    }

	if !skipTest {
		if err := testConnection(env); err != nil {
			fmt.Printf("⚠️  Connection test failed: %v\n", err)
			if !askYesNo(scanner, "Continue anyway?") {
				return fmt.Errorf("setup cancelled")
			}
		} else {
			fmt.Println("✅ Connection test successful!")
		}
	}

	if err := saveEnvironment(env, true); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	printSuccess("default")
	return nil
}

// Implements: SYS-REQ-042
func gatherEnvironmentInfo(scanner *bufio.Scanner, envName string, isFirst bool) (*types.Environment, error) {
	env := &types.Environment{Name: envName}

	fmt.Printf("📝 Configuring '%s' environment:\n", envName)
	fmt.Println()

	// Gather Dashboard URL
	if isFirst {
		fmt.Println("Enter your Tyk Dashboard URL:")
		fmt.Println("Examples:")
		fmt.Println("  • http://localhost:3000 (local development)")
		fmt.Println("  • https://admin.cloud.tyk.io (Tyk Cloud)")
		fmt.Println("  • https://dashboard.yourcompany.com (self-hosted)")
		fmt.Println()
	}
	
	env.DashboardURL = askString(scanner, "Dashboard URL", "")
	if env.DashboardURL == "" {
		return nil, fmt.Errorf("dashboard URL is required")
	}

	// Gather Auth Token
	if isFirst {
		fmt.Println("\nEnter your Dashboard API Auth Token:")
		fmt.Println("💡 You can find this in your Tyk Dashboard under 'Users' → your user → 'API Access Credentials'")
		fmt.Println()
	}
	
	env.AuthToken = askString(scanner, "Auth Token", "")
	if env.AuthToken == "" {
		return nil, fmt.Errorf("auth token is required")
	}

	// Gather Org ID
	if isFirst {
		fmt.Println("\nEnter your Organization ID:")
		fmt.Println("💡 You can find this in your Dashboard URL or in the Dashboard under 'System Management'")
		fmt.Println()
	}
	
	env.OrgID = askString(scanner, "Organization ID", "")
	if env.OrgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	return env, nil
}

// Implements: SYS-REQ-042
func testConnection(env *types.Environment) error {
	config := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": env,
		},
	}

	client, err := client.NewClient(config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Health(ctx)
}

// Implements: SYS-REQ-042
func saveEnvironment(env *types.Environment, setAsGlobal bool) error {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "cli.toml")
	
	// Create config manager and load existing config if it exists
	manager := config.NewManager()
	if _, err := os.Stat(configFile); err == nil {
		if err := manager.LoadConfig(); err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}
	}

	// Add the environment
	if err := manager.SaveEnvironment(env, setAsGlobal); err != nil { //mcdc:ignore Manager.SaveEnvironment unconditionally returns nil (config.go:144); it has no error path
		return err
	}

	// Generate and save the updated TOML config
	cfg := manager.GetConfig()
	content := generateTOMLConfigUnified(cfg)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		return err
	}

	return nil
}

// Implements: SYS-REQ-042
func printSuccess(activeEnv string) {
    fmt.Println("🎉 Setup Complete!")
    fmt.Println("==================")
    fmt.Println()
    fmt.Printf("✅ Active environment: %s\n", activeEnv)
    fmt.Println("✅ Configuration saved")
    fmt.Println()
    fmt.Println("🚀 You're ready to go! Try these commands:")
    fmt.Println("   tyk config current                # View your configuration")
    fmt.Println("   tyk api --help                    # Explore API commands")
    fmt.Println("   tyk --version                     # Check CLI version")
    fmt.Println()
    fmt.Println("📚 For more help: tyk --help")
    fmt.Println()
}

// Implements: SYS-REQ-042
func askString(scanner *bufio.Scanner, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	
	if input == "" && defaultValue != "" {
		return defaultValue
	}
	
	return input
}

// Implements: SYS-REQ-042
func askYesNo(scanner *bufio.Scanner, prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	scanner.Scan()
	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return input == "y" || input == "yes"
}

