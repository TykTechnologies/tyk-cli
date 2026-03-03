package cli

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/client"
	"github.com/tyktech/tyk-cli/internal/filehandler"
	"github.com/tyktech/tyk-cli/internal/oas"
    "github.com/tyktech/tyk-cli/pkg/types"
    "golang.org/x/term"
    "gopkg.in/yaml.v3"
)

// truncateWithEllipsis shortens s to fit max characters, adding "..." when needed
func truncateWithEllipsis(s string, max int) string {
    if max <= 0 {
        return ""
    }
    if len(s) <= max {
        return s
    }
    if max <= 3 {
        // not enough room for meaningful content
        return s[:max]
    }
    return s[:max-3] + "..."
}

// computeTableLayout returns column widths for ID/Name/Path and whether to use a stacked fallback.
func computeTableLayout(termWidth int) (idW, nameW, pathW int, stacked bool) {
    if termWidth < 20 {
        return 0, 0, 0, true
    }

    const sepWidth = 6 // two " | " separators
    contentWidth := termWidth - sepWidth
    if contentWidth < 15 {
        return 0, 0, 0, true
    }

    // Minimums and pleasant defaults
    minID, minName, minPath := 12, 14, 10
    idW, nameW, pathW = 16, 20, 14
    baseTotal := idW + nameW + pathW

    if contentWidth < (minID + minName + minPath) {
        return 0, 0, 0, true
    }

    if contentWidth < baseTotal {
        // shrink in order: id -> name -> path
        over := baseTotal - contentWidth
        shrink := func(cur *int, min int, want int) {
            if over <= 0 {
                return
            }
            can := *cur - min
            if can <= 0 {
                return
            }
            delta := want
            if delta > can {
                delta = can
            }
            if delta > over {
                delta = over
            }
            *cur -= delta
            over -= delta
        }

        shrink(&idW, minID, 8)
        shrink(&nameW, minName, 8)
        shrink(&pathW, minPath, 8)

        if over > 0 {
            return 0, 0, 0, true
        }
    } else if contentWidth > baseTotal {
        // distribute extra space conservatively to keep deterministic truncation
        extra := contentWidth - baseTotal
        grow := func(cur *int, cap int) {
            if extra <= 0 {
                return
            }
            take := cap
            if take > extra {
                take = extra
            }
            *cur += take
            extra -= take
        }
        // Limit Name to 24, Path to 16 at 80 cols (ID stays 16)
        grow(&nameW, 4)  // 20 -> up to 24
        grow(&idW, 0)    // keep at 16
        grow(&pathW, 2)  // 14 -> up to 16
    }

    return idW, nameW, pathW, false
}

func hideCursor(w io.Writer) { fmt.Fprint(w, "\x1b[?25l") }
func showCursor(w io.Writer) { fmt.Fprint(w, "\x1b[?25h") }

// readKey reads a single key or interprets ESC [ C/D as right/left arrows.
// It returns 'R' for right, 'L' for left, or the raw byte for other keys.
func readKey(r io.Reader) (byte, error) {
    buf := make([]byte, 1)
    if _, err := os.Stdin.Read(buf); err != nil { // use stdin directly (raw mode)
        return 0, err
    }
    b := buf[0]
    if b != 27 { // not ESC
        return b, nil
    }
    time.Sleep(2 * time.Millisecond)
    tail := make([]byte, 2)
    n, _ := os.Stdin.Read(tail)
    if n == 2 && tail[0] == '[' {
        switch tail[1] {
        case 'C':
            return 'R', nil // Right
        case 'D':
            return 'L', nil // Left
        }
    }
    return 27, nil // plain ESC
}

// alPrintf writes at column 0 for the current line to avoid drift from prior content.
func alPrintf(w io.Writer, format string, a ...interface{}) {
    fmt.Fprint(w, "\x1b[0G")
    fmt.Fprintf(w, format, a...)
}

// NewAPICommand creates the 'tyk api' command and its subcommands
func NewAPICommand() *cobra.Command {
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Manage OAS APIs",
		Long:  "Commands for managing OAS-native APIs in Tyk Dashboard",
	}

	// Add API subcommands
	apiCmd.AddCommand(NewAPIListCommand())
	apiCmd.AddCommand(NewAPIGetCommand())
	apiCmd.AddCommand(NewAPICreateCommand())
	apiCmd.AddCommand(NewAPIImportOASCommand())
	apiCmd.AddCommand(NewAPIApplyCommand())
	apiCmd.AddCommand(NewAPIUpdateOASCommand())
	apiCmd.AddCommand(NewAPIDeleteCommand())
	apiCmd.AddCommand(NewAPIVersionsCommand())

	return apiCmd
}

// NewAPIVersionsCommand creates the 'tyk api versions' command and its subcommands
func NewAPIVersionsCommand() *cobra.Command {
	versionsCmd := &cobra.Command{
		Use:   "versions",
		Short: "Manage API versions",
		Long:  "Commands for managing versions of OAS APIs",
	}

	// Add version subcommands
	versionsCmd.AddCommand(NewAPIVersionsListCommand())
	versionsCmd.AddCommand(NewAPIVersionsCreateCommand())
	versionsCmd.AddCommand(NewAPIVersionsSwitchDefaultCommand())

	return versionsCmd
}

// Placeholder functions for version commands - these will be implemented in phase 3

func NewAPIVersionsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List versions for an API",
		Long:  "List all versions for a given API ID",
		RunE:  runAPIVersionsList,
	}
	cmd.Flags().String("api-id", "", "API ID to list versions for (required)")
	_ = cmd.MarkFlagRequired("api-id")
	return cmd
}

func runAPIVersionsList(cmd *cobra.Command, args []string) error {
	apiID, _ := cmd.Flags().GetString("api-id")

	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	versions, defaultVersion, err := c.ListOASAPIVersions(ctx, apiID)
	if err != nil {
		if er, ok := err.(*types.ErrorResponse); ok && er.Status == 404 {
			return apiNotFoundForVersionError(apiID)
		}
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to list versions: %s", err)}
	}

	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		payload := map[string]interface{}{
			"api_id":   apiID,
			"default":  defaultVersion,
			"versions": versions,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}

	// Human output to stderr
	for _, v := range versions {
		marker := "  "
		if v == defaultVersion {
			marker = "* "
		}
		fmt.Fprintf(os.Stderr, "%s%s\n", marker, v)
	}
	fmt.Fprintf(os.Stderr, "\n%d version(s) found. Default: %s\n", len(versions), defaultVersion)
	return nil
}

func NewAPIVersionsCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new API version",
		Long:  "Create a new version for an existing API",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("API versions create command will be implemented in phase 3")
		},
	}
}

func NewAPIVersionsSwitchDefaultCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch-default",
		Short: "Switch default version for an API",
		Long:  "Switch the default version for a given API",
		RunE:  runAPIVersionsSwitchDefault,
	}
	cmd.Flags().String("api-id", "", "API ID (required)")
	cmd.Flags().String("version-name", "", "Version to set as default (required)")
	_ = cmd.MarkFlagRequired("api-id")
	_ = cmd.MarkFlagRequired("version-name")
	return cmd
}

func runAPIVersionsSwitchDefault(cmd *cobra.Command, args []string) error {
	apiID, _ := cmd.Flags().GetString("api-id")
	versionName, _ := cmd.Flags().GetString("version-name")

	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Pre-check: list versions to validate and get current default
	versions, currentDefault, err := c.ListOASAPIVersions(ctx, apiID)
	if err != nil {
		if er, ok := err.(*types.ErrorResponse); ok && er.Status == 404 {
			return apiNotFoundForVersionError(apiID)
		}
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to list versions: %s", err)}
	}

	// Check if version exists
	found := false
	for _, v := range versions {
		if v == versionName {
			found = true
			break
		}
	}
	if !found {
		return versionNotFoundError(apiID, "", versionName, versions)
	}

	outputFormat := GetOutputFormatFromContext(cmd.Context())

	// No-op if already default
	if currentDefault == versionName {
		if outputFormat == types.OutputJSON {
			payload := map[string]interface{}{
				"action":           "no_change",
				"api_id":           apiID,
				"previous_default": currentDefault,
				"new_default":      versionName,
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(payload)
		}
		fmt.Fprintf(os.Stderr, "Version %q is already the default for API %s.\n", versionName, apiID)
		return nil
	}

	// Switch
	if err := c.SwitchDefaultVersion(ctx, apiID, versionName); err != nil {
		return &ExitError{Code: 1, Message: fmt.Sprintf("failed to switch default version: %s", err)}
	}

	if outputFormat == types.OutputJSON {
		payload := map[string]interface{}{
			"action":           "switched",
			"api_id":           apiID,
			"previous_default": currentDefault,
			"new_default":      versionName,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}

	fmt.Fprintf(os.Stderr, "Previous default: %s\nNew default: %s\n", currentDefault, versionName)
	return nil
}

// Placeholder functions for API commands - these will be implemented in the next phases

func NewAPICreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API from scratch",
		Long: `Create a new OAS API from scratch with minimal configuration.

This command creates a new API by specifying basic parameters like name and upstream URL.
It auto-generates a minimal OpenAPI specification and Tyk extensions with sensible defaults.

Examples:
  tyk api create --name "User Service" --upstream-url https://users.api.com
  tyk api create --name "Payment API" --upstream-url https://payments.internal \
    --listen-path /payments/v2 --custom-domain api.company.com
  tyk api create --name "Analytics API" --upstream-url https://analytics.service \
    --description "Customer analytics and reporting" --version-name v2

After creation, you can:
  tyk api get <api-id>                           # View full configuration
  tyk api get <api-id> --oas-only > api.yaml    # Export for editing
  tyk api update-oas <api-id> --file api.yaml   # Update with enhanced spec`,
		RunE: runAPICreate,
	}

	cmd.Flags().StringP("name", "n", "", "API name (required)")
	cmd.Flags().String("upstream-url", "", "Upstream service URL (required)")
	cmd.Flags().String("listen-path", "", "API listen path (auto-generated from name if not provided)")
	cmd.Flags().String("version-name", "v1", "Version name for the API")
	cmd.Flags().String("custom-domain", "", "Custom domain for the API")
	cmd.Flags().String("description", "", "API description")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("upstream-url")

	return cmd
}

func NewAPIGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <api-id>",
		Short: "Get an API by ID",
		Long: `Retrieve an OAS API by its ID, optionally specifying a version.

By default, returns the full API metadata including Tyk-specific extensions.
Use --oas-only to get a clean OpenAPI specification without Tyk extensions,
suitable for use with standard OpenAPI tooling.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runAPIGet,
	}

	cmd.Flags().String("version-name", "", "Specific version name to retrieve")
	cmd.Flags().Bool("oas-only", false, "Return only the OpenAPI specification without Tyk extensions")

	return cmd
}

func NewAPIImportOASCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-oas",
		Short: "Import clean OpenAPI spec to create new API",
		Long: `Import a clean OpenAPI specification to create a new API.

This command imports external API specifications and creates new APIs with
automatically generated Tyk extensions. Always creates a new API ID.

Supports:
- Local files: --file petstore.yaml
- Remote URLs: --url https://api.example.com/openapi.json

For Tyk-enhanced OAS files, use 'tyk api apply' instead.`,
		RunE: runAPIImportOAS,
	}

	cmd.Flags().StringP("file", "f", "", "Path to OpenAPI specification file")
	cmd.Flags().String("url", "", "URL to OpenAPI specification")

	return cmd
}

func NewAPIApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply Tyk-enhanced API configuration",
		Long: `Apply Tyk-enhanced API configuration from an OAS file with GitOps-style declarative logic.

This command is designed for infrastructure-as-code workflows and requires files with 
x-tyk-api-gateway extensions (Tyk-enhanced OAS files).

    Behavior:
    - If x-tyk-api-gateway.info.id is present: UPSERT (update if exists, otherwise create with same ID)
    - If x-tyk-api-gateway.info.id is missing: CREATE new API

For clean OpenAPI specs without Tyk extensions, use:
- 'tyk api import-oas' to create new APIs
- 'tyk api update-oas <api-id>' to update existing APIs

Examples:
  tyk api apply --file enhanced-api.yaml    # Idempotent upsert`,
		RunE: runAPIApply,
	}

	cmd.Flags().StringP("file", "f", "", "Path to Tyk-enhanced OpenAPI specification file (use '-' for stdin) (required)")
    cmd.Flags().String("version-name", "", "Version name (defaults to info.version or v1)")
    cmd.Flags().Bool("set-default", true, "Set this version as the default")

	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func NewAPIUpdateOASCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-oas <api-id>",
		Short: "Update existing API's OpenAPI spec only",
		Long: `Update an existing API's OpenAPI specification while preserving Tyk configuration.

This command updates only the OAS portion of an existing API, preserving all
Tyk-specific middleware and configuration. It takes a clean OpenAPI spec and
merges it with existing Tyk extensions.

Supports:
- Local files: --file new-spec.yaml
- Remote URLs: --url https://api.example.com/openapi.json

For full API updates including Tyk config, use 'tyk api apply' instead.`,
		Args: cobra.ExactArgs(1),
		RunE: runAPIUpdateOAS,
	}

	cmd.Flags().StringP("file", "f", "", "Path to OpenAPI specification file")
	cmd.Flags().String("url", "", "URL to OpenAPI specification")

	return cmd
}

func NewAPIDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <api-id>",
		Short: "Delete an API by ID",
		Long:  "Delete an OAS API by its ID with confirmation prompt",
		Args:  cobra.ExactArgs(1),
		RunE:  runAPIDelete,
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	return cmd
}

// NewAPIListCommand creates the 'tyk api list' command
func NewAPIListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OAS APIs",
		Long:  "List OAS APIs in the Dashboard, paginated with optional interactive navigation",
		RunE:  runAPIList,
	}

	cmd.Flags().Int("page", 1, "Page number (10 per page)")
	cmd.Flags().BoolP("interactive", "i", false, "Enable interactive pagination with arrow key navigation")

	return cmd
}

// runAPIList implements the 'tyk api list' command
func runAPIList(cmd *cobra.Command, args []string) error {
	page, _ := cmd.Flags().GetInt("page")
	interactive, _ := cmd.Flags().GetBool("interactive")
	
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

	// If interactive mode is requested, switch to interactive pagination
	if interactive {
		if outputFormat == types.OutputJSON {
			return fmt.Errorf("interactive mode is not compatible with JSON output format")
		}
		return runInteractiveAPIList(c, page)
	}

    // Non-interactive mode (existing behavior)
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

    // Use dashboard aggregate endpoint for broader compatibility in CLI
    apis, err := c.ListAPIsDashboard(ctx, page)
	if err != nil {
		return fmt.Errorf("failed to list APIs: %w", err)
	}

	if outputFormat == types.OutputJSON {
		payload := map[string]interface{}{
			"page":  page,
			"count": len(apis),
			"apis":  apis,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}

	// Human readable output
	displayAPIPage(apis, page, false)
	return nil
}

// displayAPIPage displays a page of APIs in a formatted table
func displayAPIPage(apis []*types.OASAPI, page int, interactive bool) {
	if len(apis) == 0 {
		if interactive {
			fmt.Fprintf(os.Stderr, "\033[2J\033[H")
			fmt.Fprintf(os.Stderr, "No APIs found on page %d.\n", page)
			fmt.Fprintf(os.Stderr, "\nNavigation:\n")
			fmt.Fprintf(os.Stderr, "  ← → or A D    Previous/Next page\n")
			fmt.Fprintf(os.Stderr, "  q or Ctrl+C   Quit\n")
			fmt.Fprintf(os.Stderr, "  r             Refresh current page\n")
			fmt.Fprintf(os.Stderr, "\nPress a key to navigate... ")
		} else {
			fmt.Fprintf(os.Stderr, "No APIs found on page %d.\n", page)
		}
		return
	}

    if interactive {
        // Clear screen and move cursor to home
        fmt.Fprintf(os.Stderr, "\033[2J\033[H")

        // Determine terminal width (fallback to 80)
        termWidth := 80
        if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 {
            termWidth = w
        }
        idW, nameW, pathW, stacked := computeTableLayout(termWidth)

        // Fixed header width for consistent test expectations
        fixedHeader := 80
        alPrintf(os.Stderr, "%s\n", strings.Repeat("=", fixedHeader))
        color.New(color.FgBlue, color.Bold).Fprintf(os.Stderr, "APIs (page %d)\n", page)
        alPrintf(os.Stderr, "%s\n\n", strings.Repeat("=", fixedHeader))

        if stacked {
            for _, api := range apis {
                // Do not truncate the API ID or listen path
                alPrintf(os.Stderr, "ID: %s\n", api.ID)
                alPrintf(os.Stderr, "Name: %s\n", truncateWithEllipsis(api.Name, 48))
                alPrintf(os.Stderr, "Listen Path: %s\n", api.ListenPath)
                alPrintf(os.Stderr, "%s\n", strings.Repeat("-", 32))
            }
        } else {
            // Table header and divider with color
            hdr := color.New(color.FgCyan, color.Bold)
            dim := color.New(color.FgHiBlack)
            headerLine := fmt.Sprintf("%-*s | %-*s | %-*s", idW, "ID", nameW, "Name", pathW, "Listen Path")
            fmt.Fprint(os.Stderr, "\x1b[0G")
            hdr.Fprintln(os.Stderr, headerLine)
            dividerLine := fmt.Sprintf("%s | %s | %s", strings.Repeat("-", idW), strings.Repeat("-", nameW), strings.Repeat("-", pathW))
            fmt.Fprint(os.Stderr, "\x1b[0G")
            dim.Fprintln(os.Stderr, dividerLine)

            // Rows
            for _, api := range apis {
                // Do not truncate the API ID or listen path
                id := api.ID
                name := truncateWithEllipsis(api.Name, nameW)
                listenPath := api.ListenPath
                alPrintf(os.Stderr, "%-*s | %-*s | %-*s\n", idW, id, nameW, name, pathW, listenPath)
            }
        }

        dim := color.New(color.FgHiBlack)
        alPrintf(os.Stderr, "\n%s\n", strings.Repeat("=", fixedHeader))
        fmt.Fprint(os.Stderr, "\x1b[0G")
        dim.Fprintln(os.Stderr, "Navigation: [←→ or AD] Next/Prev | [R] Refresh | [Q] Quit")
        alPrintf(os.Stderr, "%s\n", strings.Repeat("=", fixedHeader))
        fmt.Fprint(os.Stderr, "\x1b[0G")
        dim.Fprint(os.Stderr, "Press a key to navigate... ")
    } else {
        // Non-interactive mode with colors
        blue := color.New(color.FgBlue, color.Bold)
        green := color.New(color.FgGreen, color.Bold)
		
		blue.Fprintf(os.Stderr, "APIs (page %d):\n", page)
		fmt.Fprintf(os.Stdout, "%-36s  %-28s  %-18s  %s\n", "ID", "Name", "Listen Path", "Default Version")
		fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat("-", 36+2+28+2+18+2+16))
		for _, api := range apis {
			fmt.Fprintf(os.Stdout, "%-36s  %-28s  %-18s  %s\n", api.ID, api.Name, api.ListenPath, api.DefaultVersion)
		}
		green.Fprintf(os.Stderr, "\nUse '--page %d' for next page.\n", page+1)
	}
}

// runInteractiveAPIList handles the interactive pagination mode
func runInteractiveAPIList(c *client.Client, startPage int) error {
    // Make sure we're in a terminal that supports interactive input
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return fmt.Errorf("interactive mode requires a terminal")
    }

    // Put terminal in raw mode to capture individual keystrokes
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        return fmt.Errorf("failed to enable raw terminal mode: %w", err)
    }
    defer func() {
        _ = term.Restore(int(os.Stdin.Fd()), oldState)
        showCursor(os.Stderr)
    }()

    // Hide cursor during interactive repainting
    hideCursor(os.Stderr)

	currentPage := startPage
	
	for {
		// Create context with timeout for each API call
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        // Use dashboard endpoint for interactive listing as well
        apis, err := c.ListAPIsDashboard(ctx, currentPage)
		cancel()
		
		if err != nil {
			return fmt.Errorf("failed to list APIs: %w", err)
		}

		// Display current page
		displayAPIPage(apis, currentPage, true)

        // Read a single keystroke (robust arrow handling)
        key, err := readKey(os.Stdin)
        if err != nil {
            return fmt.Errorf("failed to read input: %w", err)
        }

        switch key {
        case 'q', 'Q', 3: // 'q', 'Q', or Ctrl+C
            fmt.Fprintln(os.Stderr, "\nExiting...")
            return nil
        case 'r', 'R':
            // Refresh current page (continue loop)
            continue
        case 'a', 'A', 'L': // previous page
            if currentPage > 1 {
                currentPage--
            }
        case 'd', 'D': // next page
            // Next page - check if there are APIs on current page
            if len(apis) > 0 {
                currentPage++
            }
        default:
			// Ignore other keys
			continue
		}
	}
}

// runAPIGet implements the 'tyk api get' command
func runAPIGet(cmd *cobra.Command, args []string) error {
	apiID := args[0]
	versionName, _ := cmd.Flags().GetString("version-name")
	oasOnly, _ := cmd.Flags().GetBool("oas-only")

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

	// Get the API
	api, err := c.GetOASAPI(ctx, apiID, versionName)
	if err != nil {
		// Check if it's a not found error
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return &ExitError{Code: 3, Message: fmt.Sprintf("API '%s' not found", apiID)}
		}
		return fmt.Errorf("failed to get API: %w", err)
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		return outputAPIAsJSON(api, oasOnly)
	}

	return outputAPIAsHuman(api, versionName, oasOnly)
}

// outputAPIAsJSON outputs the API in JSON format
func outputAPIAsJSON(api *types.OASAPI, oasOnly bool) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	
	if oasOnly && api.OAS != nil {
		// Strip the x-tyk-api-gateway extension and return only the OAS
		oasData := make(map[string]interface{})
		for key, value := range api.OAS {
			if key != "x-tyk-api-gateway" {
				oasData[key] = value
			}
		}
		return encoder.Encode(oasData)
	}
	
	return encoder.Encode(api)
}

// outputAPIAsHuman outputs the API in human-readable format
func outputAPIAsHuman(api *types.OASAPI, requestedVersion string, oasOnly bool) error {
	if api == nil {
		return fmt.Errorf("API data is nil")
	}

	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)

	// Skip API summary if OAS-only mode is requested
	if !oasOnly {
		// API Summary - output to stderr so stdout can be cleanly redirected
		blue.Fprintln(os.Stderr, "API Summary:")
		fmt.Fprintf(os.Stderr, "  ID:             %s\n", api.ID)
		fmt.Fprintf(os.Stderr, "  Name:           %s\n", api.Name)
		fmt.Fprintf(os.Stderr, "  Listen Path:    %s\n", api.ListenPath)
		fmt.Fprintf(os.Stderr, "  Default Version: ")
		green.Fprintf(os.Stderr, "%s\n", api.DefaultVersion)

		if api.CustomDomain != "" {
			fmt.Fprintf(os.Stderr, "  Custom Domain:  %s\n", api.CustomDomain)
		}
		if api.UpstreamURL != "" {
			fmt.Fprintf(os.Stderr, "  Upstream URL:   %s\n", api.UpstreamURL)
		}

		fmt.Fprintf(os.Stderr, "  Created:        %s\n", api.CreatedAt)
		fmt.Fprintf(os.Stderr, "  Updated:        %s\n", api.UpdatedAt)

		// Versions summary
		if len(api.VersionData) > 0 {
			fmt.Fprintln(os.Stderr)
			blue.Fprintln(os.Stderr, "Available Versions:")
			for versionName := range api.VersionData {
				marker := ""
				if versionName == api.DefaultVersion {
					marker = green.Sprint(" (default)")
				}
				fmt.Fprintf(os.Stderr, "  - %s%s\n", versionName, marker)
			}
		}

		fmt.Fprintln(os.Stderr)
	}

	// Determine which OAS to show
	var oasData map[string]interface{}
	var versionToShow string

	if requestedVersion != "" {
		// Show specific version if requested and exists
		if versionData, exists := api.VersionData[requestedVersion]; exists && versionData.OAS != nil {
			oasData = versionData.OAS
			versionToShow = requestedVersion
		} else if api.OAS != nil {
			// Fallback to main OAS if version not found
			oasData = api.OAS
			versionToShow = "main"
			if !oasOnly {
				yellow.Fprintf(os.Stderr, "Warning: Version '%s' not found, showing main OAS document\n\n", requestedVersion)
			}
		}
	} else {
		// No specific version requested, show main OAS
		oasData = api.OAS
		versionToShow = "main"
	}

	if oasData != nil {
		// Strip x-tyk-api-gateway extension if OAS-only mode is requested
		if oasOnly {
			filteredOAS := make(map[string]interface{})
			for key, value := range oasData {
				if key != "x-tyk-api-gateway" {
					filteredOAS[key] = value
				}
			}
			oasData = filteredOAS
		} else {
			// Header to stderr (only in non-OAS-only mode)
			blue.Fprintf(os.Stderr, "OpenAPI Specification")
			if versionToShow != "main" {
				blue.Fprintf(os.Stderr, " (version: %s)", versionToShow)
			}
			blue.Fprintln(os.Stderr, ":")
		}

		// Convert to YAML for better readability and output to stdout
		yamlData, err := yaml.Marshal(oasData)
		if err != nil {
			return fmt.Errorf("failed to convert OAS to YAML: %w", err)
		}

		// Output YAML to stdout (no color for clean piping)
		fmt.Print(string(yamlData))
	} else {
		if !oasOnly {
			yellow.Fprintln(os.Stderr, "No OAS document available")
		}
	}

	return nil
}

// runAPIImportOAS implements the 'tyk api import-oas' command
func runAPIImportOAS(cmd *cobra.Command, args []string) error {
	// Get flags
	filePath, _ := cmd.Flags().GetString("file")
	urlFlag, _ := cmd.Flags().GetString("url")

	// Validate input: either file or url must be provided
	if filePath == "" && urlFlag == "" {
		return &ExitError{Code: 2, Message: "Either --file or --url must be provided"}
	}
	if filePath != "" && urlFlag != "" {
		return &ExitError{Code: 2, Message: "Cannot specify both --file and --url"}
	}

	// Get configuration from context
	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	// Load OAS data from file or URL
	var oasData map[string]interface{}
	var err error

	if filePath != "" {
		// Load from file
		oasData, err = loadOASFromFile(filePath)
	} else {
		// Load from URL
		oasData, err = loadOASFromURL(urlFlag)
	}
	if err != nil {
		return err
	}

	// Auto-generate x-tyk-api-gateway extensions for plain OAS documents
	if !oas.HasTykExtensions(oasData) {
		oasData, err = oas.AddTykExtensions(oasData)
		if err != nil {
			return &ExitError{Code: 2, Message: fmt.Sprintf("failed to generate Tyk extensions: %v", err)}
		}
	}

	// Strip any existing API ID from OAS file (import always generates new ID)
	oasData = stripExistingAPIID(oasData)

	// Extract version name from OAS document
	versionName := extractVersionFromOAS(oasData)
	if versionName == "" {
		versionName = "v1" // fallback
	}

	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the API
	api, err := c.CreateOASAPI(ctx, oasData)
	if err != nil {
		// Check for conflict errors
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "conflict") {
			return &ExitError{Code: 4, Message: fmt.Sprintf("API import failed due to conflict: %v", err)}
		}
		return fmt.Errorf("failed to import API: %w", err)
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		return outputImportedAPIAsJSON(api, versionName)
	}

	return outputImportedAPIAsHuman(api, versionName)
}

// extractVersionFromOAS extracts version from OAS info.version field
func extractVersionFromOAS(oasData map[string]interface{}) string {
	if info, ok := oasData["info"].(map[string]interface{}); ok {
		if version, ok := info["version"].(string); ok && version != "" {
			return version
		}
	}
	return ""
}

// outputImportedAPIAsJSON outputs the imported API result in JSON format
func outputImportedAPIAsJSON(api *types.OASAPI, versionName string) error {
	result := map[string]interface{}{
		"api_id":          api.ID,
		"version_name":    versionName,
		"name":            api.Name,
		"listen_path":     api.ListenPath,
		"default_version": api.DefaultVersion,
		"operation":       "imported",
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// outputImportedAPIAsHuman outputs the imported API result in human-readable format
func outputImportedAPIAsHuman(api *types.OASAPI, versionName string) error {
	green := color.New(color.FgGreen, color.Bold)
	blue := color.New(color.FgBlue, color.Bold)

	green.Println("✓ API imported successfully!")
	fmt.Printf("  API ID:         %s\n", api.ID)
	fmt.Printf("  Name:           %s\n", api.Name)
	fmt.Printf("  Version:        %s\n", versionName)
	fmt.Printf("  Listen Path:    %s\n", api.ListenPath)

	if api.CustomDomain != "" {
		fmt.Printf("  Custom Domain:  %s\n", api.CustomDomain)
	}
	if api.UpstreamURL != "" {
		fmt.Printf("  Upstream URL:   %s\n", api.UpstreamURL)
	}

	blue.Printf("  Default Version: %s\n", api.DefaultVersion)

	return nil
}

// stripExistingAPIID removes x-tyk-api-gateway.info.id from OAS document
// This ensures create command always generates new ID
func stripExistingAPIID(oasData map[string]interface{}) map[string]interface{} {
	if xTyk, exists := oasData["x-tyk-api-gateway"]; exists {
		if xTykMap, ok := xTyk.(map[string]interface{}); ok {
			if info, exists := xTykMap["info"]; exists {
				if infoMap, ok := info.(map[string]interface{}); ok {
					delete(infoMap, "id") // Remove existing API ID
				}
			}
		}
	}
	return oasData
}

// runAPIApply implements the 'tyk api apply' command (declarative upsert)
func runAPIApply(cmd *cobra.Command, args []string) error {
    // Get flags
    filePath, _ := cmd.Flags().GetString("file")
    versionName, _ := cmd.Flags().GetString("version-name")
    setDefault, _ := cmd.Flags().GetBool("set-default")

	// Get configuration from context
	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

    var oasData map[string]interface{}
    if filePath == "-" {
        // Read from stdin; support JSON or YAML (YAML parser also accepts JSON)
        data, err := io.ReadAll(os.Stdin)
        if err != nil {
            return &ExitError{Code: 2, Message: fmt.Sprintf("failed to read stdin: %v", err)}
        }
        if len(data) == 0 {
            return &ExitError{Code: 2, Message: "no input provided on stdin"}
        }
        if err := yaml.Unmarshal(data, &oasData); err != nil {
            return &ExitError{Code: 2, Message: fmt.Sprintf("failed to parse input as YAML/JSON: %v", err)}
        }
    } else {
        // Validate and read the OAS file
        if !filepath.IsAbs(filePath) {
            absPath, err := filepath.Abs(filePath)
            if err != nil {
                return &ExitError{Code: 2, Message: fmt.Sprintf("failed to resolve file path: %v", err)}
            }
            filePath = absPath
        }

        // Check if file exists
        if _, err := os.Stat(filePath); os.IsNotExist(err) {
            return &ExitError{Code: 2, Message: fmt.Sprintf("file not found: %s", filePath)}
        }

        // Load and parse the OAS file
        fileInfo, err := filehandler.LoadFile(filePath)
        if err != nil {
            return &ExitError{Code: 2, Message: fmt.Sprintf("failed to load OAS file: %v", err)}
        }
        oasData = fileInfo.Content
    }

	// Enhanced validation: Check if it's a Tyk-enhanced OAS file
    if !oas.HasTykExtensions(oasData) {
        return &ExitError{
            Code:    2,
            Message: "File lacks required x-tyk-api-gateway extensions. This command requires Tyk-enhanced OAS files.\n\nFor clean OpenAPI specs, use:\n  tyk api import-oas --file " + filepath.Base(filePath) + "  # To create new API\n  tyk api update-oas <api-id> --file " + filepath.Base(filePath) + "  # To update existing API",
        }
    }

	// Check for existing API ID in the file
	apiID, hasID := oas.ExtractAPIIDFromTykExtensions(oasData)

    if hasID {
        // API ID present - upsert (update or create if missing)
        return updateExistingAPI(cmd, config, apiID, oasData, versionName, setDefault)
    }

    // No API ID present - create new API automatically
    return createNewAPIViaApply(cmd, config, oasData, versionName, setDefault)
}

// updateExistingAPI handles updating an existing API via apply
func updateExistingAPI(cmd *cobra.Command, config *types.Config, apiID string, oasData map[string]interface{}, versionName string, setDefault bool) error {
	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

    // Check if API exists first. If not found, create it with the same ID (idempotent upsert)
    _, err = c.GetOASAPI(ctx, apiID, "")
    if err != nil {
        // Determine if the error means "not found" for upsert semantics
        notFound := false
        if er, ok := err.(*types.ErrorResponse); ok {
            // Treat 404 as not found; also handle some Dashboard variants that return 400 for missing IDs
            if er.Status == 404 {
                notFound = true
            } else if er.Status == 400 {
                msg := strings.ToLower(er.Message)
                if strings.Contains(msg, "could not retrieve api") || strings.Contains(msg, "not found") {
                    notFound = true
                }
            }
        } else if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
            notFound = true
        }

        if notFound {
            // Fallback to create with provided ID in the OAS
            if versionName == "" {
                versionName = extractVersionFromOAS(oasData)
                if versionName == "" {
                    versionName = "v1"
                }
            }

            api, cerr := c.CreateOASAPI(ctx, oasData)
            if cerr != nil {
                if strings.Contains(cerr.Error(), "409") || strings.Contains(cerr.Error(), "conflict") {
                    return &ExitError{Code: 4, Message: fmt.Sprintf("API creation failed due to conflict: %v", cerr)}
                }
                return fmt.Errorf("failed to create API: %w", cerr)
            }

            // Output creation result
            outputFormat := GetOutputFormatFromContext(cmd.Context())
            if outputFormat == types.OutputJSON {
                return outputImportedAPIAsJSON(api, versionName)
            }
            return outputImportedAPIAsHuman(api, versionName)
        }

        return fmt.Errorf("failed to verify API exists: %w", err)
    }

	// Extract version name from OAS if not provided
	if versionName == "" {
		versionName = extractVersionFromOAS(oasData)
		if versionName == "" {
			versionName = "v1" // fallback
		}
	}

	// Update the API
	api, err := c.UpdateOASAPI(ctx, apiID, oasData)
	if err != nil {
		return fmt.Errorf("failed to update API: %w", err)
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		return outputUpdatedAPIAsJSON(api, versionName)
	}

	return outputUpdatedAPIAsHuman(api, versionName)
}

// createNewAPIViaApply handles creating a new API via apply
func createNewAPIViaApply(cmd *cobra.Command, config *types.Config, oasData map[string]interface{}, versionName string, setDefault bool) error {
	// Auto-generate x-tyk-api-gateway extensions for plain OAS documents
	if !oas.HasTykExtensions(oasData) {
		var err error
		oasData, err = oas.AddTykExtensions(oasData)
		if err != nil {
			return &ExitError{Code: 2, Message: fmt.Sprintf("failed to generate Tyk extensions: %v", err)}
		}
	}

	// Strip any existing ID (shouldn't be there, but be safe)
	oasData = stripExistingAPIID(oasData)

	// Extract version name from OAS if not provided
	if versionName == "" {
		versionName = extractVersionFromOAS(oasData)
		if versionName == "" {
			versionName = "v1" // fallback
		}
	}

	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the API
	api, err := c.CreateOASAPI(ctx, oasData)
	if err != nil {
		// Check for conflict errors
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "conflict") {
			return &ExitError{Code: 4, Message: fmt.Sprintf("API creation failed due to conflict: %v", err)}
		}
		return fmt.Errorf("failed to create API: %w", err)
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		return outputImportedAPIAsJSON(api, versionName)
	}

	return outputImportedAPIAsHuman(api, versionName)
}

// runAPIUpdateOAS implements the 'tyk api update-oas' command
func runAPIUpdateOAS(cmd *cobra.Command, args []string) error {
	// Get API ID from args
	apiID := args[0]

	// Get flags
	filePath, _ := cmd.Flags().GetString("file")
	urlFlag, _ := cmd.Flags().GetString("url")

	// Validate input: either file or url must be provided
	if filePath == "" && urlFlag == "" {
		return &ExitError{Code: 2, Message: "Either --file or --url must be provided"}
	}
	if filePath != "" && urlFlag != "" {
		return &ExitError{Code: 2, Message: "Cannot specify both --file and --url"}
	}

	// Get configuration from context
	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	// Load OAS data from file or URL
	var oasData map[string]interface{}
	var err error

	if filePath != "" {
		// Load from file
		oasData, err = loadOASFromFile(filePath)
	} else {
		// Load from URL
		oasData, err = loadOASFromURL(urlFlag)
	}
	if err != nil {
		return err
	}

	return updateExistingAPIWithOAS(cmd, config, apiID, oasData)
}

// runAPIDelete implements the 'tyk api delete' command
func runAPIDelete(cmd *cobra.Command, args []string) error {
	apiID := args[0]
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

	// Verify API exists first
	api, err := c.GetOASAPI(ctx, apiID, "")
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return &ExitError{Code: 3, Message: fmt.Sprintf("API '%s' not found", apiID)}
		}
		return fmt.Errorf("failed to verify API exists: %w", err)
	}

	// Confirmation prompt unless --yes flag is provided
	if !skipConfirmation {
		fmt.Printf("Are you sure you want to delete API '%s' (%s)? [y/N]: ", apiID, api.Name)
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Delete operation cancelled")
			return nil
		}
	}

	// Delete the API
	err = c.DeleteOASAPI(ctx, apiID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return &ExitError{Code: 3, Message: fmt.Sprintf("API '%s' not found", apiID)}
		}
		return fmt.Errorf("failed to delete API: %w", err)
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		return outputDeletedAPIAsJSON(apiID)
	}

	return outputDeletedAPIAsHuman(apiID, api.Name)
}

// outputUpdatedAPIAsJSON outputs the updated API result in JSON format
func outputUpdatedAPIAsJSON(api *types.OASAPI, versionName string) error {
	result := map[string]interface{}{
		"api_id":          api.ID,
		"version_name":    versionName,
		"name":            api.Name,
		"listen_path":     api.ListenPath,
		"default_version": api.DefaultVersion,
		"operation":       "updated",
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// outputUpdatedAPIAsHuman outputs the updated API result in human-readable format
func outputUpdatedAPIAsHuman(api *types.OASAPI, versionName string) error {
	green := color.New(color.FgGreen, color.Bold)
	blue := color.New(color.FgBlue, color.Bold)

	green.Println("✓ API updated successfully!")
	fmt.Printf("  API ID:         %s\n", api.ID)
	fmt.Printf("  Name:           %s\n", api.Name)
	fmt.Printf("  Version:        %s\n", versionName)
	fmt.Printf("  Listen Path:    %s\n", api.ListenPath)

	if api.CustomDomain != "" {
		fmt.Printf("  Custom Domain:  %s\n", api.CustomDomain)
	}
	if api.UpstreamURL != "" {
		fmt.Printf("  Upstream URL:   %s\n", api.UpstreamURL)
	}

	blue.Printf("  Default Version: %s\n", api.DefaultVersion)

	return nil
}

// outputDeletedAPIAsJSON outputs the deleted API result in JSON format
func outputDeletedAPIAsJSON(apiID string) error {
	result := map[string]interface{}{
		"api_id":    apiID,
		"operation": "deleted",
		"success":   true,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// outputDeletedAPIAsHuman outputs the deleted API result in human-readable format
func outputDeletedAPIAsHuman(apiID, apiName string) error {
	green := color.New(color.FgGreen, color.Bold)

	green.Printf("✓ Deleted API '%s'\n", apiID)
	if apiName != "" {
		fmt.Printf("  Name: %s\n", apiName)
	}

	return nil
}

// loadOASFromFile loads and parses an OAS file from the local filesystem
func loadOASFromFile(filePath string) (map[string]interface{}, error) {
	// Validate and read the OAS file
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return nil, &ExitError{Code: 2, Message: fmt.Sprintf("failed to resolve file path: %v", err)}
		}
		filePath = absPath
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, &ExitError{Code: 2, Message: fmt.Sprintf("file not found: %s", filePath)}
	}

	// Load and parse the OAS file
	fileInfo, err := filehandler.LoadFile(filePath)
	if err != nil {
		return nil, &ExitError{Code: 2, Message: fmt.Sprintf("failed to load OAS file: %v", err)}
	}

	return fileInfo.Content, nil
}

// loadOASFromURL loads and parses an OAS document from a URL
func loadOASFromURL(urlStr string) (map[string]interface{}, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch the URL
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, &ExitError{Code: 2, Message: fmt.Sprintf("failed to fetch URL: %v", err)}
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, &ExitError{Code: 2, Message: fmt.Sprintf("failed to fetch URL: HTTP %d", resp.StatusCode)}
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ExitError{Code: 2, Message: fmt.Sprintf("failed to read URL response: %v", err)}
	}

	// Parse as JSON or YAML
	var oasData map[string]interface{}
	
	// Try JSON first
	if err := json.Unmarshal(body, &oasData); err != nil {
		// Try YAML
		if err := yaml.Unmarshal(body, &oasData); err != nil {
			return nil, &ExitError{Code: 2, Message: fmt.Sprintf("failed to parse OAS document: %v", err)}
		}
	}

	return oasData, nil
}

// runAPICreate implements the 'tyk api create' command
func runAPICreate(cmd *cobra.Command, args []string) error {
	// Get flags
	name, _ := cmd.Flags().GetString("name")
	upstreamURL, _ := cmd.Flags().GetString("upstream-url")
	listenPath, _ := cmd.Flags().GetString("listen-path")
	versionName, _ := cmd.Flags().GetString("version-name")
	customDomain, _ := cmd.Flags().GetString("custom-domain")
	description, _ := cmd.Flags().GetString("description")

	// Auto-generate listen path if not provided
	if listenPath == "" {
		listenPath = oas.GenerateListenPath(name)
	}

	// Set default description if not provided
	if description == "" {
		description = "Auto-generated API specification"
	}

	// Get configuration from context
	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	// Generate the OAS document with Tyk extensions
	oasData, err := generateOASForCreate(name, description, versionName, upstreamURL, listenPath, customDomain)
	if err != nil {
		return fmt.Errorf("failed to generate OAS document: %w", err)
	}

	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the API
	api, err := c.CreateOASAPI(ctx, oasData)
	if err != nil {
		// Check for conflict errors
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "conflict") {
			return &ExitError{Code: 4, Message: fmt.Sprintf("API creation failed due to conflict: %v", err)}
		}
		return fmt.Errorf("failed to create API: %w", err)
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		return outputCreatedAPIAsJSON(api, versionName)
	}

	return outputCreatedAPIAsHuman(api, versionName)
}

// generateOASForCreate creates a minimal OAS document with Tyk extensions for the create command
func generateOASForCreate(name, description, version, upstreamURL, listenPath, customDomain string) (map[string]interface{}, error) {
	// Create basic OAS structure
	oasDoc := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       name,
			"description": description,
			"version":     version,
		},
		"servers": []interface{}{
			map[string]interface{}{
				"url": upstreamURL,
			},
		},
		"paths": map[string]interface{}{},
	}

	// Add Tyk extensions
	tykExtensions := map[string]interface{}{
		"info": map[string]interface{}{
			"name": name,
			"state": map[string]interface{}{
				"active": true,
			},
		},
		"upstream": map[string]interface{}{
			"url": upstreamURL,
		},
		"server": map[string]interface{}{
			"listenPath": map[string]interface{}{
				"value": listenPath,
				"strip": true,
			},
		},
	}

	// Add custom domain if provided
	if customDomain != "" {
		tykExtensions["server"].(map[string]interface{})["customDomain"] = map[string]interface{}{
			"enabled": true,
			"name":    customDomain,
		}
	}

	oasDoc["x-tyk-api-gateway"] = tykExtensions

	return oasDoc, nil
}

// outputCreatedAPIAsJSON outputs the created API result in JSON format
func outputCreatedAPIAsJSON(api *types.OASAPI, versionName string) error {
	result := map[string]interface{}{
		"api_id":          api.ID,
		"version_name":    versionName,
		"name":            api.Name,
		"listen_path":     api.ListenPath,
		"default_version": api.DefaultVersion,
		"operation":       "created",
	}

	if api.CustomDomain != "" {
		result["custom_domain"] = api.CustomDomain
	}
	if api.UpstreamURL != "" {
		result["upstream_url"] = api.UpstreamURL
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// outputCreatedAPIAsHuman outputs the created API result in human-readable format
func outputCreatedAPIAsHuman(api *types.OASAPI, versionName string) error {
	green := color.New(color.FgGreen, color.Bold)
	blue := color.New(color.FgBlue, color.Bold)
	dim := color.New(color.FgHiBlack)

	green.Println("✓ API created successfully!")
	fmt.Printf("  API ID:         %s\n", api.ID)
	fmt.Printf("  Name:           %s\n", api.Name)
	fmt.Printf("  Version:        %s\n", versionName)
	fmt.Printf("  Listen Path:    %s\n", api.ListenPath)

	if api.CustomDomain != "" {
		fmt.Printf("  Custom Domain:  %s\n", api.CustomDomain)
	}
	if api.UpstreamURL != "" {
		fmt.Printf("  Upstream URL:   %s\n", api.UpstreamURL)
	}

	blue.Printf("  Default Version: %s\n", api.DefaultVersion)

	fmt.Printf("\n")
	dim.Println("Next steps:")
	dim.Printf("  tyk api get %s                           # View full configuration\n", api.ID)
	dim.Printf("  tyk api get %s --oas-only > api.yaml    # Export for editing\n", api.ID)

	return nil
}

// updateExistingAPIWithOAS handles updating an existing API with a clean OAS document
func updateExistingAPIWithOAS(cmd *cobra.Command, config *types.Config, apiID string, oasData map[string]interface{}) error {
	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if API exists first and get existing Tyk extensions
	existingAPI, err := c.GetOASAPI(ctx, apiID, "")
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return &ExitError{Code: 3, Message: fmt.Sprintf("API with ID '%s' not found", apiID)}
		}
		return fmt.Errorf("failed to verify API exists: %w", err)
	}

	// Preserve existing Tyk extensions by merging with new OAS
	if existingAPI.OAS != nil {
		if tykExt, exists := existingAPI.OAS["x-tyk-api-gateway"]; exists {
			oasData["x-tyk-api-gateway"] = tykExt
		}
	}

	// If no Tyk extensions found, generate them for the clean OAS
	if !oas.HasTykExtensions(oasData) {
		oasData, err = oas.AddTykExtensions(oasData)
		if err != nil {
			return &ExitError{Code: 2, Message: fmt.Sprintf("failed to generate Tyk extensions: %v", err)}
		}
	}

	// Ensure the API ID matches in the extensions
	if tykExt, exists := oasData["x-tyk-api-gateway"]; exists {
		if tykExtMap, ok := tykExt.(map[string]interface{}); ok {
			if info, exists := tykExtMap["info"]; exists {
				if infoMap, ok := info.(map[string]interface{}); ok {
					infoMap["id"] = apiID
				}
			}
		}
	}

	// Extract version name from OAS document
	versionName := extractVersionFromOAS(oasData)
	if versionName == "" {
		versionName = "v1" // fallback
	}

	// Update the API
	api, err := c.UpdateOASAPI(ctx, apiID, oasData)
	if err != nil {
		return fmt.Errorf("failed to update API: %w", err)
	}

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	if outputFormat == types.OutputJSON {
		return outputUpdatedAPIAsJSON(api, versionName)
	}

	return outputUpdatedAPIAsHuman(api, versionName)
}
