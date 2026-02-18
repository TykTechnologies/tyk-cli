package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/client"
	"github.com/tyktech/tyk-cli/internal/policy"
	"github.com/tyktech/tyk-cli/pkg/types"
	"gopkg.in/yaml.v3"
)

// NewPolicyCommand creates the 'tyk policy' command and its subcommands
func NewPolicyCommand() *cobra.Command {
	policyCmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage policies",
		Long:  "Commands for managing security policies in Tyk Dashboard",
	}

	policyCmd.AddCommand(NewPolicyListCommand())
	policyCmd.AddCommand(NewPolicyGetCommand())
	policyCmd.AddCommand(NewPolicyApplyCommand())
	policyCmd.AddCommand(NewPolicyDeleteCommand())
	policyCmd.AddCommand(NewPolicyInitCommand())

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

// NewPolicyGetCommand creates the 'tyk policy get' command
func NewPolicyGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <policy-id>",
		Short: "Get a policy by ID",
		Long:  "Retrieve a security policy by ID, convert to CLI schema, and output as YAML or JSON",
		Args:  cobra.ExactArgs(1),
		RunE:  runPolicyGet,
	}

	return cmd
}

// runPolicyGet implements the 'tyk policy get' command
func runPolicyGet(cmd *cobra.Command, args []string) error {
	policyID := args[0]

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

	// Fetch the policy
	dp, err := c.GetPolicy(ctx, policyID)
	if err != nil {
		if er, ok := err.(*types.ErrorResponse); ok && er.Status == 404 {
			return &ExitError{Code: int(types.ExitNotFound), Message: fmt.Sprintf("policy '%s' not found", policyID)}
		}
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return &ExitError{Code: int(types.ExitNotFound), Message: fmt.Sprintf("policy '%s' not found", policyID)}
		}
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to get policy: %v", err)}
	}

	// Fetch API list for reverse-resolution of API IDs to names
	apis, err := c.ListAPIsDashboard(ctx, 1)
	if err != nil {
		// Non-fatal: proceed without reverse resolution
		apis = nil
	}

	// Convert OAS APIs to ResolverAPI for WireToCLI
	resolverAPIs := make([]policy.ResolverAPI, 0, len(apis))
	for _, api := range apis {
		resolverAPIs = append(resolverAPIs, policy.ResolverAPI{
			ID:         api.ID,
			Name:       api.Name,
			ListenPath: api.ListenPath,
		})
	}

	// Convert wire format to CLI schema
	pf := policy.WireToCLI(*dp, resolverAPIs)

	if outputFormat == types.OutputJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(pf)
	}

	// Human mode: summary to stderr, YAML to stdout
	fmt.Fprintf(os.Stderr, "Policy: %s\n", pf.Metadata.Name)
	fmt.Fprintf(os.Stderr, "  ID:   %s\n", pf.Metadata.ID)
	if len(pf.Metadata.Tags) > 0 {
		fmt.Fprintf(os.Stderr, "  Tags: %s\n", strings.Join(pf.Metadata.Tags, ", "))
	}
	apiCount := len(pf.Spec.Access)
	fmt.Fprintf(os.Stderr, "  APIs: %d\n", apiCount)

	yamlData, err := yaml.Marshal(pf)
	if err != nil {
		return fmt.Errorf("failed to marshal policy as YAML: %w", err)
	}
	fmt.Fprint(os.Stdout, string(yamlData))

	return nil
}

// NewPolicyApplyCommand creates the 'tyk policy apply' command
func NewPolicyApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a policy from a YAML file",
		Long: `Apply a policy from a YAML file with idempotent upsert semantics.

Creates the policy if metadata.id is not found on the server; updates if it exists.

Selectors in spec.access resolve against the live API list:
  - name: exact match on API name
  - listenPath: exact match on listen path
  - id: exact match on API ID
  - tags: expand to all APIs matching all specified tags

Examples:
  tyk policy apply -f policy.yaml         # Apply from file
  cat policy.yaml | tyk policy apply -f - # Apply from stdin`,
		RunE: runPolicyApply,
	}

	cmd.Flags().StringP("file", "f", "", "Path to policy YAML file (use '-' for stdin) (required)")
	cmd.MarkFlagRequired("file")

	return cmd
}

// runPolicyApply implements the 'tyk policy apply' command
func runPolicyApply(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")

	// Get configuration from context
	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	// Step 1: Read YAML from file or stdin
	var data []byte
	var err error

	if filePath == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return &ExitError{Code: int(types.ExitBadArgs), Message: fmt.Sprintf("failed to read stdin: %v", err)}
		}
		if len(data) == 0 {
			return &ExitError{Code: int(types.ExitBadArgs), Message: "no input provided on stdin"}
		}
	} else {
		data, err = os.ReadFile(filePath)
		if err != nil {
			return &ExitError{Code: int(types.ExitBadArgs), Message: fmt.Sprintf("failed to read file: %v", err)}
		}
	}

	// Step 2: Unmarshal YAML to PolicyFile
	var pf types.PolicyFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return &ExitError{Code: int(types.ExitBadArgs), Message: fmt.Sprintf("failed to parse YAML: %v", err)}
	}

	// Step 3: Validate schema
	if validationErrs := policy.ValidatePolicy(pf); len(validationErrs) > 0 {
		var msgs []string
		for _, ve := range validationErrs {
			msgs = append(msgs, ve.Error())
		}
		return &ExitError{Code: int(types.ExitBadArgs), Message: strings.Join(msgs, "; ")}
	}

	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 4: Fetch API list from Dashboard
	apis, err := c.ListAPIsDashboard(ctx, 1)
	if err != nil {
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to fetch API list: %v", err)}
	}

	// Step 5: Convert to ResolverAPI slice
	resolverAPIs := make([]policy.ResolverAPI, 0, len(apis))
	for _, api := range apis {
		resolverAPIs = append(resolverAPIs, policy.ResolverAPI{
			ID:         api.ID,
			Name:       api.Name,
			ListenPath: api.ListenPath,
		})
	}

	// Step 6: Build resolve requests from access entries and resolve selectors
	requests := make([]policy.ResolveRequest, 0, len(pf.Spec.Access))
	for _, entry := range pf.Spec.Access {
		req := policy.ResolveRequest{Versions: entry.Versions}
		switch {
		case entry.ID != "":
			req.SelectorType = "id"
			req.Value = entry.ID
		case entry.Name != "":
			req.SelectorType = "name"
			req.Value = entry.Name
		case entry.ListenPath != "":
			req.SelectorType = "listenPath"
			req.Value = entry.ListenPath
		case len(entry.Tags) > 0:
			req.SelectorType = "tags"
			req.TagValues = entry.Tags
		}
		requests = append(requests, req)
	}

	resolved, resolveErrs := policy.ResolveAccessEntries(requests, resolverAPIs)
	if len(resolveErrs) > 0 {
		var msgs []string
		for _, re := range resolveErrs {
			msgs = append(msgs, re.Error())
		}
		return &ExitError{Code: int(types.ExitBadArgs), Message: strings.Join(msgs, "; ")}
	}

	// Step 7: Convert CLI to wire format
	activeEnv, err := config.GetActiveEnvironment()
	if err != nil {
		return fmt.Errorf("no active environment: %w", err)
	}
	dp, err := policy.CLIToWire(pf, resolved, activeEnv.OrgID)
	if err != nil {
		return &ExitError{Code: int(types.ExitBadArgs), Message: err.Error()}
	}

	// Step 8: Check if policy exists
	_, getErr := c.GetPolicy(ctx, pf.Metadata.ID)

	policyExists := false
	if getErr == nil {
		policyExists = true
	} else {
		// Check if it's a "not found" error
		notFound := false
		if er, ok := getErr.(*types.ErrorResponse); ok && er.Status == 404 {
			notFound = true
		} else if strings.Contains(getErr.Error(), "404") || strings.Contains(strings.ToLower(getErr.Error()), "not found") {
			notFound = true
		}
		if !notFound {
			return &ExitError{Code: 1, Message: fmt.Sprintf("failed to check existing policy: %v", getErr)}
		}
	}

	// Step 9: Create or Update
	if policyExists {
		if err := c.UpdatePolicy(ctx, pf.Metadata.ID, &dp); err != nil {
			return &ExitError{Code: 1, Message: fmt.Sprintf("failed to update policy: %v", err)}
		}
		fmt.Fprintf(os.Stderr, "Policy '%s' (%s) updated.\n", pf.Metadata.Name, pf.Metadata.ID)
	} else {
		if err := c.CreatePolicy(ctx, &dp); err != nil {
			return &ExitError{Code: 1, Message: fmt.Sprintf("failed to create policy: %v", err)}
		}
		fmt.Fprintf(os.Stderr, "Policy '%s' (%s) created.\n", pf.Metadata.Name, pf.Metadata.ID)
	}

	return nil
}

// NewPolicyDeleteCommand creates the 'tyk policy delete' command
func NewPolicyDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <policy-id>",
		Short: "Delete a policy by ID",
		Long:  "Delete a security policy by ID with confirmation prompt",
		Args:  cobra.ExactArgs(1),
		RunE:  runPolicyDelete,
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	return cmd
}

// runPolicyDelete implements the 'tyk policy delete' command
func runPolicyDelete(cmd *cobra.Command, args []string) error {
	policyID := args[0]
	skipConfirmation, _ := cmd.Flags().GetBool("yes")

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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch policy to verify existence and get name for confirmation
	dp, err := c.GetPolicy(ctx, policyID)
	if err != nil {
		if er, ok := err.(*types.ErrorResponse); ok && er.Status == 404 {
			return &ExitError{Code: int(types.ExitNotFound), Message: fmt.Sprintf("policy '%s' not found", policyID)}
		}
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return &ExitError{Code: int(types.ExitNotFound), Message: fmt.Sprintf("policy '%s' not found", policyID)}
		}
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to get policy: %v", err)}
	}

	// Confirmation prompt unless --yes flag is provided
	if !skipConfirmation {
		fmt.Fprintf(os.Stderr, "Are you sure you want to delete policy '%s' (%s)? [y/N]: ", dp.Name, policyID)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Fprintf(os.Stderr, "Delete operation cancelled.\n")
			return nil
		}
	}

	// Delete the policy
	if err := c.DeletePolicy(ctx, policyID); err != nil {
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to delete policy: %v", err)}
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		result := map[string]interface{}{
			"policy_id": policyID,
			"operation": "deleted",
			"success":   true,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	// Human-readable confirmation to stderr
	fmt.Fprintf(os.Stderr, "Policy '%s' (%s) deleted.\n", dp.Name, policyID)
	return nil
}

// NewPolicyInitCommand creates the 'tyk policy init' command
func NewPolicyInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a scaffold policy YAML file",
		Long: `Generate a scaffold policy YAML file with sensible defaults.

If --id and --name are provided, prompts are skipped (non-interactive mode).
The file is written to policies/{id}.yaml relative to --dir (default: current directory).
Refuses to overwrite an existing file.

Examples:
  tyk policy init --id my-policy --name "My Policy"
  tyk policy init --id gold --name "Gold Plan" --dir ./project`,
		RunE: runPolicyInit,
	}

	cmd.Flags().String("id", "", "Policy ID (required)")
	cmd.Flags().String("name", "", "Policy name (required)")
	cmd.Flags().String("dir", ".", "Base directory for output (policies/{id}.yaml is created inside)")

	return cmd
}

// runPolicyInit implements the 'tyk policy init' command
func runPolicyInit(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetString("id")
	name, _ := cmd.Flags().GetString("name")
	dir, _ := cmd.Flags().GetString("dir")

	if id == "" {
		return &ExitError{Code: int(types.ExitBadArgs), Message: "policy ID is required (use --id)"}
	}
	if name == "" {
		return &ExitError{Code: int(types.ExitBadArgs), Message: "policy name is required (use --name)"}
	}

	// Build output path
	policiesDir := filepath.Join(dir, "policies")
	outPath := filepath.Join(policiesDir, id+".yaml")

	// Check if file already exists
	if _, err := os.Stat(outPath); err == nil {
		return &ExitError{Code: int(types.ExitBadArgs), Message: fmt.Sprintf("file already exists: %s", outPath)}
	}

	// Generate scaffold
	pf := types.PolicyFile{
		APIVersion: "tyk.tyktech/v1",
		Kind:       "Policy",
		Metadata: types.PolicyMetadata{
			ID:   id,
			Name: name,
		},
		Spec: types.PolicySpec{
			RateLimit: &types.RateLimit{
				Requests: 1000,
				Per:      types.Duration("1m"),
			},
			Quota: &types.Quota{
				Limit:  100000,
				Period: types.Duration("30d"),
			},
			KeyTTL: types.Duration("0"),
			Access: []types.AccessEntry{
				{
					Name:     "your-api-name",
					Versions: []string{"Default"},
				},
			},
		},
	}

	data, err := yaml.Marshal(pf)
	if err != nil {
		return fmt.Errorf("failed to marshal scaffold: %w", err)
	}

	// Ensure policies directory exists
	if err := os.MkdirAll(policiesDir, 0755); err != nil {
		return fmt.Errorf("failed to create policies directory: %w", err)
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write scaffold file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Policy scaffold written to %s\n", outPath)
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
