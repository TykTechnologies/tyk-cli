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

const httpTimeout = 30 * time.Second

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
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	// Resolve friendly ID to fetch the policy (O(1) GET)
	dp, err := resolveFriendlyID(ctx, c, policyID)
	if err != nil {
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to resolve policy: %v", err)}
	}
	if dp == nil {
		return &ExitError{Code: int(types.ExitNotFound), Message: fmt.Sprintf("policy '%s' not found", policyID)}
	}

	// Fetch API list for reverse-resolution of API IDs to names (non-fatal on error)
	apis, err := c.ListAPIsDashboard(ctx, 1)
	if err != nil {
		apis = nil
	}

	resolverAPIs := toResolverAPIs(apis)

	// Convert wire format to CLI schema
	pf := policy.WireToCLI(*dp, resolverAPIs)

	if outputFormat == types.OutputJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(pf)
	}

	// Human mode: summary to stderr, YAML to stdout
	fmt.Fprintf(os.Stderr, "Policy: %s\n", pf.Name)
	fmt.Fprintf(os.Stderr, "  ID:   %s\n", pf.ID)
	if len(pf.Tags) > 0 {
		fmt.Fprintf(os.Stderr, "  Tags: %s\n", strings.Join(pf.Tags, ", "))
	}
	apiCount := len(pf.Access)
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
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// runPolicyApply implements the 'tyk policy apply' command
func runPolicyApply(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")

	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	pf, err := readPolicyFile(filePath)
	if err != nil {
		return err
	}

	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	// Fetch API list and resolve selectors
	apis, err := c.ListAPIsDashboard(ctx, 1)
	if err != nil {
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to fetch API list: %v", err)}
	}

	requests := buildResolveRequests(pf.Access)
	resolved, resolveErrs := policy.ResolveAccessEntries(requests, toResolverAPIs(apis))
	if len(resolveErrs) > 0 {
		return &ExitError{Code: int(types.ExitBadArgs), Message: joinErrorMessages(resolveErrs)}
	}

	// Convert CLI to wire format
	activeEnv, err := config.GetActiveEnvironment()
	if err != nil {
		return fmt.Errorf("no active environment: %w", err)
	}
	dp, err := policy.CLIToWire(pf, resolved, activeEnv.OrgID)
	if err != nil {
		return &ExitError{Code: int(types.ExitBadArgs), Message: err.Error()}
	}

	// Resolve friendly ID to check if policy already exists (O(1) GET)
	existingPolicy, resolveErr := resolveFriendlyID(ctx, c, pf.ID)
	if resolveErr != nil {
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to resolve policy: %v", resolveErr)}
	}

	// Create or update based on existence check
	if existingPolicy != nil {
		// Update path — use the friendly ID
		if err := c.UpdatePolicy(ctx, pf.ID, &dp); err != nil {
			return &ExitError{Code: 1, Message: fmt.Sprintf("failed to update policy: %v", err)}
		}
		fmt.Fprintf(os.Stderr, "Policy '%s' (%s) updated.\n", pf.Name, pf.ID)
	} else {
		// Create path — omit _id, let Dashboard generate it
		dp.MID = ""
		if err := c.CreatePolicy(ctx, &dp); err != nil {
			return &ExitError{Code: 1, Message: fmt.Sprintf("failed to create policy: %v", err)}
		}
		fmt.Fprintf(os.Stderr, "Policy '%s' (%s) created.\n", pf.Name, pf.ID)
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
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	// Resolve friendly ID to fetch the policy (O(1) GET)
	dp, err := resolveFriendlyID(ctx, c, policyID)
	if err != nil {
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to resolve policy: %v", err)}
	}
	if dp == nil {
		return &ExitError{Code: int(types.ExitNotFound), Message: fmt.Sprintf("policy '%s' not found", policyID)}
	}

	// Confirmation prompt unless --yes flag is provided
	if !skipConfirmation {
		fmt.Fprintf(os.Stderr, "Are you sure you want to delete policy '%s' (%s)? [y/N]: ", dp.Name, policyID)
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Fprintf(os.Stderr, "Delete operation cancelled.\n")
			return nil
		}
	}

	// Delete the policy using the friendly ID
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
		ID:   id,
		Name: name,
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

// readPolicyFile reads a policy YAML from a file path or stdin ("-"), parses it,
// and validates the schema. Returns the parsed PolicyFile or an error.
func readPolicyFile(filePath string) (types.PolicyFile, error) {
	var data []byte
	var err error

	if filePath == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return types.PolicyFile{}, &ExitError{Code: int(types.ExitBadArgs), Message: fmt.Sprintf("failed to read stdin: %v", err)}
		}
		if len(data) == 0 {
			return types.PolicyFile{}, &ExitError{Code: int(types.ExitBadArgs), Message: "no input provided on stdin"}
		}
	} else {
		data, err = os.ReadFile(filePath)
		if err != nil {
			return types.PolicyFile{}, &ExitError{Code: int(types.ExitBadArgs), Message: fmt.Sprintf("failed to read file: %v", err)}
		}
	}

	var pf types.PolicyFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return types.PolicyFile{}, &ExitError{Code: int(types.ExitBadArgs), Message: fmt.Sprintf("failed to parse YAML: %v", err)}
	}

	if validationErrs := policy.ValidatePolicy(pf); len(validationErrs) > 0 {
		msgs := make([]string, len(validationErrs))
		for i := range validationErrs {
			msgs[i] = validationErrs[i].Error()
		}
		return types.PolicyFile{}, &ExitError{Code: int(types.ExitBadArgs), Message: strings.Join(msgs, "; ")}
	}

	return pf, nil
}

// buildResolveRequests converts access entries from a PolicyFile into ResolveRequests.
func buildResolveRequests(entries []types.AccessEntry) []policy.ResolveRequest {
	requests := make([]policy.ResolveRequest, 0, len(entries))
	for _, entry := range entries {
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
	return requests
}

// joinErrorMessages concatenates error messages into a semicolon-separated string.
func joinErrorMessages(errs []error) string {
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

func resolveFriendlyID(ctx context.Context, c *client.Client, friendlyID string) (*types.DashboardPolicy, error) {
	dp, err := c.GetPolicy(ctx, friendlyID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil // not found — caller decides create or error
		}
		return nil, fmt.Errorf("failed to resolve policy %q: %w", friendlyID, err)
	}
	return dp, nil
}

// isNotFoundError returns true if the error indicates a 404 / not found response.
func isNotFoundError(err error) bool {
	if er, ok := err.(*types.ErrorResponse); ok && er.Status == 404 {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "404") || strings.Contains(strings.ToLower(msg), "not found")
}

// toResolverAPIs converts OAS API objects to the resolver's input type.
func toResolverAPIs(apis []*types.OASAPI) []policy.ResolverAPI {
	result := make([]policy.ResolverAPI, 0, len(apis))
	for _, api := range apis {
		result = append(result, policy.ResolverAPI{
			ID:         api.ID,
			Name:       api.Name,
			ListenPath: api.ListenPath,
		})
	}
	return result
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
		displayID := p.ID
		if displayID == "" {
			displayID = p.MID // unmanaged policy — show _id
		}
		apiCount := len(p.AccessRights)
		tags := strings.Join(p.Tags, ", ")
		fmt.Fprintf(os.Stdout, "%-26s  %-24s  %-10d  %s\n", displayID, p.Name, apiCount, tags)
	}
	fmt.Fprintf(os.Stderr, "\nUse '--page %d' for next page.\n", page+1)
}
