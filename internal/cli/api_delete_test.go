package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// reqproof:req REQ-API-007
func executeAPIDeleteCmd(t *testing.T, serverURL string, outputFormat types.OutputFormat, apiID string, yes bool) error {
	t.Helper()
	deleteCmd := NewAPIDeleteCommand()

	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: serverURL, AuthToken: "token", OrgID: "org"},
		},
	}
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, outputFormat)
	deleteCmd.SetContext(ctx)

	cmdArgs := []string{apiID}
	if yes {
		cmdArgs = append(cmdArgs, "--yes")
	}
	deleteCmd.SetArgs(cmdArgs)
	_ = deleteCmd.ParseFlags(cmdArgs)

	return deleteCmd.RunE(deleteCmd, []string{apiID})
}

// reqproof:req REQ-API-007
func TestAPIDelete_WithYes(t *testing.T) {
	deleteCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apis/oas/test-api-id":
			deleteCalled = true
			_ = json.NewEncoder(w).Encode(types.APIResponse{Status: "success", Message: "deleted"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", true)

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)
	assert.True(t, deleteCalled, "DELETE should have been called")
	assert.Contains(t, string(stdout), "Test API", "stdout should reference the deleted API name")
}

// reqproof:req REQ-API-007
func TestAPIDelete_WithYes_JSON(t *testing.T) {
	deleteCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apis/oas/test-api-id":
			deleteCalled = true
			_ = json.NewEncoder(w).Encode(types.APIResponse{Status: "success", Message: "deleted"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executeAPIDeleteCmd(t, server.URL, types.OutputJSON, "test-api-id", true)

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)
	assert.True(t, deleteCalled, "DELETE should have been called")

	var result map[string]interface{}
	err = json.Unmarshal(stdout, &result)
	require.NoError(t, err, "output should be valid JSON")
	assert.Equal(t, "test-api-id", result["api_id"])
	assert.Equal(t, "deleted", result["operation"])
}

// reqproof:req REQ-API-007
func TestAPIDelete_NotFound_OnVerify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "API not found"})
	}))
	defer server.Close()

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "nonexistent", true)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError, got %T: %v", err, err)
	assert.Equal(t, 3, exitErr.Code)
	assert.Contains(t, exitErr.Message, "not found")
}

// reqproof:req REQ-API-007
func TestAPIDelete_NotFound_OnDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apis/oas/test-api-id":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "API not found"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", true)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError, got %T: %v", err, err)
	assert.Equal(t, 3, exitErr.Code)
	assert.Contains(t, exitErr.Message, "not found")
}

// ---------------------------------------------------------------------------
// MC/DC coverage for runAPIDelete confirmation prompt and edge cases
// ---------------------------------------------------------------------------

// reqproof:req REQ-API-007
// withStdin temporarily replaces os.Stdin with a pipe containing `input` for
// the duration of `body`.
func withStdin(t *testing.T, input string, body func()) {
	t.Helper()
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.Write([]byte(input))
	w.Close()
	defer func() { os.Stdin = oldStdin }()
	body()
}

// reqproof:req REQ-API-007
// TestAPIDelete_PromptCancelled covers L1207 !skipConfirmation=T plus L1211
// response not in {y,yes}.
func TestAPIDelete_PromptCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id" {
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	var err error
	withStdin(t, "n\n", func() {
		err = executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", false)
	})
	require.NoError(t, err, "cancelling the prompt should not be an error")
}

// reqproof:req REQ-API-007
// TestAPIDelete_PromptConfirmedYes covers L1211 response=="y" (proves the
// negated AND short-circuit on the "yes" half).
func TestAPIDelete_PromptConfirmedYes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(types.APIResponse{Status: "success"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var err error
	withStdin(t, "y\n", func() {
		err = executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", false)
	})
	require.NoError(t, err)
}

// reqproof:req REQ-API-007
// TestAPIDelete_PromptConfirmedYesWord covers L1211 response=="yes" (proves
// the full-word "yes" branch).
func TestAPIDelete_PromptConfirmedYesWord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(types.APIResponse{Status: "success"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var err error
	withStdin(t, "yes\n", func() {
		err = executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", false)
	})
	require.NoError(t, err)
}

// reqproof:req REQ-API-007
// TestAPIDelete_ServerErrorOnDelete covers L1219 err!=nil after Delete with
// a non-404 wrap path (L1220 substring branches both false).
func TestAPIDelete_ServerErrorOnDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apis/oas/test-api-id":
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 500, "message": "server boom"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", true)
	require.Error(t, err)
}

// reqproof:req REQ-API-007
// TestAPIDelete_VerifyFails_NonNotFound covers L1197 substring branches both
// false (classifyDashboardError handles), and confirms 500 maps to ExitError(8).
func TestAPIDelete_VerifyFails_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 500, "message": "server boom"})
	}))
	defer server.Close()

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", true)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitServerError), exitErr.Code)
}

// reqproof:req REQ-API-007
// TestAPIDelete_DeleteNotFound_404OnlyMessage covers L1220 substring branch
// where "404" matches but "not found" does not (on the DELETE call).
func TestAPIDelete_DeleteNotFound_404OnlyMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apis/oas/test-api-id":
			// Non-JSON 404 body so the wrapped error contains "404" only.
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("api deleted already"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", true)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, 3, exitErr.Code)
}

// reqproof:req REQ-API-007
// TestAPIDelete_VerifyFails_NonClassified covers L1200 cls!=nil=F branch
// (non-404 non-classified error on GET).
func TestAPIDelete_VerifyFails_NonClassified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad"})
	}))
	defer server.Close()

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "x", true)
	require.Error(t, err)
}

// reqproof:req REQ-API-007
// TestAPIDelete_VerifyNotFound_404OnlyMessage covers the L1197 substring branch
// where "404" matches but "not found" doesn't.
func TestAPIDelete_VerifyNotFound_404OnlyMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// JSON body with controlled message that contains "404" only.
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 400, "message": "got 404 from upstream",
		})
	}))
	defer server.Close()

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "x", true)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, 3, exitErr.Code)
}

// reqproof:req REQ-API-007
// TestAPIDelete_DeleteNotFound_404Only covers L1220 substring branch where the
// DELETE wraps a message with "404" only.
func TestAPIDelete_DeleteNotFound_404Only(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/test-api-id":
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apis/oas/test-api-id":
			// JSON 400 with "404" in message; we expect substring branch T,F.
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": 400, "message": "got 404 from gateway",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executeAPIDeleteCmd(t, server.URL, types.OutputHuman, "test-api-id", true)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, 3, exitErr.Code)
}

// reqproof:req REQ-API-007
// TestAPIDelete_NewClientFails covers L1186 err!=nil from client.NewClient.
func TestAPIDelete_NewClientFails(t *testing.T) {
	cmd := NewAPIDeleteCommand()
	cmd.SetContext(withConfig(context.Background(), brokenConfig()))
	cmd.SetArgs([]string{"x", "--yes"})
	_ = cmd.ParseFlags([]string{"x", "--yes"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
}

// reqproof:req REQ-API-007
// TestAPIDelete_ConfigNil covers L1180 config==nil=T branch.
func TestAPIDelete_ConfigNil(t *testing.T) {
	cmd := NewAPIDeleteCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"x", "--yes"})
	_ = cmd.ParseFlags([]string{"x", "--yes"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not found")
}

// reqproof:req REQ-API-007
// TestAPIDelete_JSONOutput_Yes covers L1229 outputFormat==OutputJSON path.
// (Covered partially by TestAPIDelete_WithYes_JSON; explicit reqproof here.)
func TestAPIDelete_JSONOutput_PromptCancelled_NoOutput(t *testing.T) {
	// Cancelled prompt should NOT emit JSON and return nil.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	var err error
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	withStdin(t, "n\n", func() {
		err = executeAPIDeleteCmd(t, server.URL, types.OutputJSON, "test-api-id", false)
	})
	wOut.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(rOut)
	require.NoError(t, err)
	assert.NotContains(t, string(out), "\"operation\"", "cancelled delete should not emit success JSON")
}
