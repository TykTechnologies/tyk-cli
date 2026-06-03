package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reqproof:req REQ-API-040
// versionsPlaceholderCmds enumerates the three placeholder subcommands so the
// tests can iterate over them with their expected output substring.
var versionsPlaceholderCmds = []struct {
	name        string
	build       func() *placeholderCmd
	wantSubstr  string
}{
	{"list", buildList, "API versions list command will be implemented in phase 3"},
	{"create", buildCreate, "API versions create command will be implemented in phase 3"},
	{"switch-default", buildSwitchDefault, "API versions switch-default command will be implemented in phase 3"},
}

// reqproof:req REQ-API-040
type placeholderCmd struct {
	execute func(out *bytes.Buffer)
}

// reqproof:req REQ-API-040
func buildList() *placeholderCmd {
	cmd := NewAPIVersionsListCommand()
	return &placeholderCmd{execute: func(out *bytes.Buffer) {
		cmd.SetOut(out)
		cmd.Run(cmd, nil)
	}}
}

// reqproof:req REQ-API-040
func buildCreate() *placeholderCmd {
	cmd := NewAPIVersionsCreateCommand()
	return &placeholderCmd{execute: func(out *bytes.Buffer) {
		cmd.SetOut(out)
		cmd.Run(cmd, nil)
	}}
}

// reqproof:req REQ-API-040
func buildSwitchDefault() *placeholderCmd {
	cmd := NewAPIVersionsSwitchDefaultCommand()
	return &placeholderCmd{execute: func(out *bytes.Buffer) {
		cmd.SetOut(out)
		cmd.Run(cmd, nil)
	}}
}

// reqproof:req REQ-API-040
func TestAPIVersions_PlaceholderSubcommandsEmitMessage(t *testing.T) {
	for _, tc := range versionsPlaceholderCmds {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			tc.build().execute(&out)
			assert.Contains(t, out.String(), tc.wantSubstr,
				"%s placeholder must emit the documented message", tc.name)
		})
	}
}

// reqproof:req REQ-API-040
// TestAPIVersions_PlaceholderSubcommandsMakeNoDashboardCall verifies the
// placeholder commands return without any HTTP activity. They are constructed
// without a config/context, so any attempt to reach the Dashboard would panic
// on nil dereference; survival of the call proves no dashboard contact.
func TestAPIVersions_PlaceholderSubcommandsMakeNoDashboardCall(t *testing.T) {
	for _, tc := range versionsPlaceholderCmds {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			require.NotPanics(t, func() {
				tc.build().execute(&out)
			}, "%s placeholder must not contact the Dashboard (which would require config and panic without it)", tc.name)
		})
	}
}
