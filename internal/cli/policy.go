package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/client"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// NewPolicyCommand creates the 'tyk policy' command and its subcommands
func NewPolicyCommand() *cobra.Command {
	policyCmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage policies",
		Long:  "Commands for managing security policies in Tyk Dashboard",
	}

	policyCmd.AddCommand(NewPolicyListCommand())

	return policyCmd
}

// NewPolicyListCommand creates the 'tyk policy list' command
func NewPolicyListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List policies",
		Long:  "List security policies in the Dashboard, paginated",
		RunE:  runPolicyList,
	}

	cmd.Flags().Int("page", 1, "Page number")

	return cmd
}

// runPolicyList implements the 'tyk policy list' command
func runPolicyList(cmd *cobra.Command, args []string) error {
	page, _ := cmd.Flags().GetInt("page")
	if page <= 0 {
		page = 1
	}

	// Get configuration from context
	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch policies
	result, err := c.ListPolicies(ctx, page)
	if err != nil {
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to list policies: %v", err)}
	}

	policies := result.Data

	if outputFormat == types.OutputJSON {
		payload := map[string]interface{}{
			"page":     page,
			"count":    len(policies),
			"policies": policies,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}

	// Human readable output
	displayPolicyPage(policies, page)
	return nil
}

// displayPolicyPage displays a page of policies in a formatted table
func displayPolicyPage(policies []types.DashboardPolicy, page int) {
	if len(policies) == 0 {
		fmt.Fprintf(os.Stderr, "No policies found.\n")
		return
	}

	fmt.Fprintf(os.Stderr, "Policies (page %d):\n", page)
	fmt.Fprintf(os.Stdout, "%-26s  %-24s  %-10s  %s\n", "ID", "Name", "APIs", "Tags")
	fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat("-", 26+2+24+2+10+2+20))
	for _, p := range policies {
		apiCount := len(p.AccessRights)
		tags := strings.Join(p.Tags, ", ")
		fmt.Fprintf(os.Stdout, "%-26s  %-24s  %-10d  %s\n", p.MID, p.Name, apiCount, tags)
	}
	fmt.Fprintf(os.Stderr, "\nUse '--page %d' for next page.\n", page+1)
}
