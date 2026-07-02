package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/tyktech/tyk-cli/internal/cli"
)

// Build-time variables (set by ldflags)
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// classifyExitError maps an Execute() error into the process exit code,
// writing the user-facing message to errOut. Returns 0 when err is nil.
// Extracted so main()'s error-handling can be tested without spawning a
// subprocess or wrapping os.Exit.
// Implements: SYS-REQ-013
func classifyExitError(err error, errOut io.Writer) int {
	if err == nil {
		return 0
	}
	var exitError *cli.ExitError
	if errors.As(err, &exitError) {
		fmt.Fprintf(errOut, "Error: %v\n", exitError.Message)
		return exitError.Code
	}
	fmt.Fprintf(errOut, "Error: %v\n", err)
	return 1
}

// Implements: SYS-REQ-013
func main() {
	rootCmd := cli.NewRootCommand(version, commit, buildTime)
	os.Exit(classifyExitError(rootCmd.Execute(), os.Stderr))
}