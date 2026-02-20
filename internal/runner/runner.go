// Package runner executes terraform plan against each configured workspace.
package runner

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Result holds the outcome of running terraform plan in a single workspace.
type Result struct {
	// WorkspacePath is the directory of the workspace that was scanned.
	WorkspacePath string
	// PlanOutput is the raw JSON output from terraform plan -json.
	PlanOutput []byte
	// Stderr is the captured stderr from the terraform plan invocation.
	Stderr []byte
	// ExitCode is the process exit code (0=no changes, 1=error, 2=changes present).
	ExitCode int
	// Err holds any execution error (e.g., binary not found, permission denied).
	Err error
}

// Options configures the workspace runner.
type Options struct {
	// Binary is the terraform (or tofu) binary to invoke.
	// Defaults to "terraform" if empty.
	Binary string
}

// RunWorkspace executes terraform plan -json -detailed-exitcode in the given
// workspace directory and returns the result.
func RunWorkspace(workspacePath string, opts Options) Result {
	result := Result{WorkspacePath: workspacePath}

	binary := opts.Binary
	if binary == "" {
		binary = "terraform"
	}

	cmd := exec.Command(binary, "plan", "-json", "-detailed-exitcode")
	cmd.Dir = workspacePath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.PlanOutput = stdout.Bytes()
	result.Stderr = stderr.Bytes()

	if exitErr, ok := err.(*exec.ExitError); ok {
		// Command ran but exited with non-zero code
		result.ExitCode = exitErr.ExitCode()
		result.Err = nil // Exit code 2 is not an error for us, it's drift detected
	} else if err != nil {
		// Command failed to run (binary not found, etc.)
		result.Err = fmt.Errorf("running terraform plan in %s: %w", workspacePath, err)
		result.ExitCode = 2 // Treat as scan error
	} else {
		result.ExitCode = 0
	}

	return result
}

// RunAll iterates workspacePaths sequentially and returns a result per workspace.
func RunAll(workspacePaths []string, opts Options) []Result {
	results := make([]Result, 0, len(workspacePaths))
	for _, path := range workspacePaths {
		result := RunWorkspace(path, opts)
		results = append(results, result)
	}
	return results
}
