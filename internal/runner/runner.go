// Package runner executes terraform plan against each configured workspace.
package runner

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
	// TODO: implement in Task 2
	return Result{WorkspacePath: workspacePath}
}

// RunAll iterates workspacePaths sequentially and returns a result per workspace.
func RunAll(workspacePaths []string, opts Options) []Result {
	// TODO: implement in Task 2
	return nil
}
