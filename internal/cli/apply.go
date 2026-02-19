package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/oas"
	"gopkg.in/yaml.v3"
)

// configFileType represents the classified type of a config file.
type configFileType string

const (
	configFileAPI          configFileType = "api"
	configFilePolicy       configFileType = "policy"
	configFileUnrecognized configFileType = "unrecognized"
)

// discoveredFile represents a file found during directory scanning.
type discoveredFile struct {
	Path    string
	Type    configFileType
	Content map[string]interface{}
}

// NewApplyCommand creates the top-level 'tyk apply' command.
func NewApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply API and policy configurations from files or directories",
		Long: `Apply API and policy configurations from a file or directory.

Recursively discovers .yaml, .yml, and .json files, classifies them as
API definitions (x-tyk-api-gateway) or policies (id + name + access_rights),
and applies them to the Dashboard. Policies are applied before APIs.

Examples:
  tyk apply -f ./configs/          # Apply all configs in directory
  tyk apply -f api.yaml            # Apply a single file`,
		RunE: runApply,
	}

	cmd.Flags().StringP("file", "f", "", "Path to file or directory (required)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runApply(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")

	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	// Determine if path is file or directory
	info, err := os.Stat(filePath)
	if err != nil {
		return &ExitError{Code: 2, Message: fmt.Sprintf("cannot access path: %v", err)}
	}

	var files []discoveredFile
	if info.IsDir() {
		files, err = scanDirectory(filePath)
		if err != nil {
			return &ExitError{Code: 2, Message: fmt.Sprintf("failed to scan directory: %v", err)}
		}
	} else {
		f, err := classifyFile(filePath)
		if err != nil {
			return &ExitError{Code: 2, Message: fmt.Sprintf("failed to read file: %v", err)}
		}
		files = []discoveredFile{*f}
	}

	// Filter: only API and policy files count; unrecognized are kept for error reporting
	var configFiles []discoveredFile
	for _, f := range files {
		configFiles = append(configFiles, f)
	}

	// Check for empty (no config files at all)
	hasConfig := false
	for _, f := range configFiles {
		if f.Type == configFileAPI || f.Type == configFilePolicy {
			hasConfig = true
			break
		}
	}

	// If all files are unrecognized, we still have files but they are failures.
	// If there are zero files at all, that's the empty dir case.
	if len(configFiles) == 0 || !hasConfig {
		// Check if we only have unrecognized files
		if len(configFiles) > 0 {
			// We have unrecognized files - report them as failed
			// Fall through to apply logic which will report failures
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "No API or policy files found in %s\n", filePath)
			return &ExitError{Code: 2, Message: "No API or policy files found"}
		}
	}

	// Sort: policies first, then APIs, alphabetical within each group
	sortFiles(configFiles)

	// Build batch applier
	activeEnv, err := config.GetActiveEnvironment()
	if err != nil {
		return fmt.Errorf("no active environment: %w", err)
	}
	applier := &batchApplier{
		baseURL:   activeEnv.DashboardURL,
		authToken: activeEnv.AuthToken,
		orgID:     activeEnv.OrgID,
		client:    &http.Client{Timeout: 30 * time.Second},
	}

	// Apply files
	total := len(configFiles)
	applied := 0
	failed := 0
	w := cmd.ErrOrStderr()

	for i, f := range configFiles {
		baseName := filepath.Base(f.Path)

		if f.Type == configFileUnrecognized {
			fmt.Fprintf(w, "[%d/%d] %s ... FAILED (unrecognized file type)\n", i+1, total, baseName)
			failed++
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		applyErr := applier.applyFile(ctx, f)
		cancel()

		if applyErr != nil {
			fmt.Fprintf(w, "[%d/%d] %s ... FAILED (%v)\n", i+1, total, baseName, applyErr)
			failed++
			continue
		}

		applied++
		fmt.Fprintf(w, "[%d/%d] %s ... OK\n", i+1, total, baseName)
	}

	// Summary
	fmt.Fprintf(w, "\nApply complete: %d/%d succeeded", applied, total)
	if failed > 0 {
		fmt.Fprintf(w, ", %d failed", failed)
	}
	fmt.Fprintln(w)

	if failed > 0 {
		return &ExitError{Code: 1, Message: fmt.Sprintf("%d file(s) failed", failed)}
	}

	return nil
}

// scanDirectory recursively finds and classifies config files in a directory.
func scanDirectory(dir string) ([]discoveredFile, error) {
	var files []discoveredFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			// Silently skip non-config files
			return nil
		}

		f, err := classifyFile(path)
		if err != nil {
			// Parse error - mark as unrecognized
			files = append(files, discoveredFile{
				Path: path,
				Type: configFileUnrecognized,
			})
			return nil
		}

		files = append(files, *f)
		return nil
	})

	return files, err
}

// classifyFile loads and classifies a single file.
func classifyFile(path string) (*discoveredFile, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" && ext != ".json" {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var content map[string]interface{}
	if ext == ".json" {
		if err := json.Unmarshal(data, &content); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, &content); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	}

	fileType := classifyContent(content)

	return &discoveredFile{
		Path:    path,
		Type:    fileType,
		Content: content,
	}, nil
}

// classifyContent determines if parsed content is an API, policy, or unrecognized.
func classifyContent(content map[string]interface{}) configFileType {
	// API: has x-tyk-api-gateway extension
	if oas.HasTykExtensions(content) {
		return configFileAPI
	}

	// Policy: has id + name + access_rights (the heuristic from the spec)
	_, hasID := content["id"]
	_, hasName := content["name"]
	_, hasAccess := content["access_rights"]
	if hasID && hasName && hasAccess {
		return configFilePolicy
	}

	return configFileUnrecognized
}

// sortFiles sorts discovered files: policies first, then APIs, alphabetical within each.
func sortFiles(files []discoveredFile) {
	sort.SliceStable(files, func(i, j int) bool {
		// Type priority: policy < api < unrecognized
		pi := typePriority(files[i].Type)
		pj := typePriority(files[j].Type)
		if pi != pj {
			return pi < pj
		}
		return filepath.Base(files[i].Path) < filepath.Base(files[j].Path)
	})
}

func typePriority(t configFileType) int {
	switch t {
	case configFilePolicy:
		return 0
	case configFileAPI:
		return 1
	default:
		return 2
	}
}

// batchApplier handles direct HTTP communication with the Dashboard for batch operations.
type batchApplier struct {
	baseURL   string
	authToken string
	orgID     string
	client    *http.Client
}

func (ba *batchApplier) applyFile(ctx context.Context, f discoveredFile) error {
	switch f.Type {
	case configFileAPI:
		return ba.applyAPI(ctx, f)
	case configFilePolicy:
		return ba.applyPolicy(ctx, f)
	default:
		return fmt.Errorf("unrecognized file type")
	}
}

func (ba *batchApplier) applyAPI(ctx context.Context, f discoveredFile) error {
	apiID, hasID := oas.ExtractAPIIDFromTykExtensions(f.Content)

	if hasID && apiID != "" {
		// Check if API exists
		status, _ := ba.doJSON(ctx, http.MethodGet, "/api/apis/oas/"+url.PathEscape(apiID), nil)
		if status == http.StatusOK {
			// Update
			s, body := ba.doJSON(ctx, http.MethodPut, "/api/apis/oas/"+url.PathEscape(apiID), f.Content)
			if s >= 400 {
				return fmt.Errorf("update failed (%d): %s", s, body)
			}
			return nil
		}
	}

	// Create
	s, body := ba.doJSON(ctx, http.MethodPost, "/api/apis/oas", f.Content)
	if s >= 400 {
		return fmt.Errorf("create failed (%d): %s", s, body)
	}
	return nil
}

func (ba *batchApplier) applyPolicy(ctx context.Context, f discoveredFile) error {
	policyID, _ := f.Content["id"].(string)
	if policyID == "" {
		return fmt.Errorf("policy file missing 'id' field")
	}

	f.Content["org_id"] = ba.orgID

	// Check if policy exists
	status, _ := ba.doJSON(ctx, http.MethodGet, "/api/portal/policies/"+url.PathEscape(policyID), nil)
	if status == http.StatusOK {
		// Update
		s, body := ba.doJSON(ctx, http.MethodPut, "/api/portal/policies/"+url.PathEscape(policyID), f.Content)
		if s >= 400 {
			return fmt.Errorf("update failed (%d): %s", s, body)
		}
		return nil
	}

	// Create
	s, body := ba.doJSON(ctx, http.MethodPost, "/api/portal/policies", f.Content)
	if s >= 400 {
		return fmt.Errorf("create failed (%d): %s", s, body)
	}
	return nil
}

func (ba *batchApplier) doJSON(ctx context.Context, method, path string, payload interface{}) (int, string) {
	var reqBody io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, ba.baseURL+path, reqBody)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", ba.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ba.client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}
