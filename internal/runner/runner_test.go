package runner_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/daemonship/driftwatch/internal/runner"
)

func TestRunWorkspace_BinaryNotFound(t *testing.T) {
	result := runner.RunWorkspace("/tmp", runner.Options{Binary: "nonexistent-binary-xyz"})
	if result.Err == nil {
		t.Error("RunWorkspace() Err = nil, want error for missing binary")
	}
	if result.ExitCode != 2 {
		t.Errorf("RunWorkspace() ExitCode = %d, want 2 for scan error", result.ExitCode)
	}
}

func TestRunWorkspace_WorkspacePathPreserved(t *testing.T) {
	dir := t.TempDir()
	result := runner.RunWorkspace(dir, runner.Options{Binary: "nonexistent-binary-xyz"})
	if result.WorkspacePath != dir {
		t.Errorf("RunWorkspace() WorkspacePath = %q, want %q", result.WorkspacePath, dir)
	}
}

func TestRunWorkspace_DefaultBinary(t *testing.T) {
	// When Binary is empty, should default to "terraform".
	// We can't run real terraform, but we can verify the binary name attempted.
	dir := t.TempDir()
	result := runner.RunWorkspace(dir, runner.Options{})
	// Either terraform is found or not — the point is it attempted "terraform".
	// We just verify no panic and WorkspacePath is preserved.
	if result.WorkspacePath != dir {
		t.Errorf("RunWorkspace() WorkspacePath = %q, want %q", result.WorkspacePath, dir)
	}
}

func TestRunWorkspace_CapturesStderr(t *testing.T) {
	// Use a fake terraform binary that writes to stderr and exits non-zero.
	fakeTerraform := buildFakeTerraform(t, `
package main
import (
	"fmt"
	"os"
)
func main() {
	fmt.Fprintln(os.Stderr, "Error: No configuration files found")
	os.Exit(1)
}
`)
	dir := t.TempDir()
	result := runner.RunWorkspace(dir, runner.Options{Binary: fakeTerraform})
	if len(result.Stderr) == 0 {
		t.Error("RunWorkspace() Stderr is empty, want stderr captured from process")
	}
}

func TestRunWorkspace_ExitCode2MeansDrift(t *testing.T) {
	// terraform plan -detailed-exitcode exits 2 when changes are present.
	fakeTerraform := buildFakeTerraform(t, `
package main
import (
	"fmt"
	"os"
)
func main() {
	fmt.Print(`+"`"+`{"format_version":"1.2","resource_changes":[{"address":"aws_instance.web","change":{"actions":["update"],"before":{"ami":"ami-old"},"after":{"ami":"ami-new"}}}]}`+"`"+`)
	os.Exit(2)
}
`)
	dir := t.TempDir()
	result := runner.RunWorkspace(dir, runner.Options{Binary: fakeTerraform})
	if result.ExitCode != 2 {
		t.Errorf("RunWorkspace() ExitCode = %d, want 2 for drift", result.ExitCode)
	}
	if len(result.PlanOutput) == 0 {
		t.Error("RunWorkspace() PlanOutput is empty, want JSON output captured")
	}
}

func TestRunWorkspace_ExitCode0MeansNoDrift(t *testing.T) {
	fakeTerraform := buildFakeTerraform(t, `
package main
import (
	"fmt"
)
func main() {
	fmt.Print(`+"`"+`{"format_version":"1.2","resource_changes":[]}`+"`"+`)
}
`)
	dir := t.TempDir()
	result := runner.RunWorkspace(dir, runner.Options{Binary: fakeTerraform})
	if result.ExitCode != 0 {
		t.Errorf("RunWorkspace() ExitCode = %d, want 0 for no drift", result.ExitCode)
	}
	if result.Err != nil {
		t.Errorf("RunWorkspace() Err = %v, want nil for successful no-op plan", result.Err)
	}
}

func TestRunAll_ReturnsOneResultPerWorkspace(t *testing.T) {
	paths := []string{"/path/one", "/path/two", "/path/three"}
	results := runner.RunAll(paths, runner.Options{Binary: "nonexistent-binary-xyz"})
	if len(results) != len(paths) {
		t.Errorf("RunAll() returned %d results, want %d", len(results), len(paths))
	}
	for i, r := range results {
		if r.WorkspacePath != paths[i] {
			t.Errorf("RunAll() result[%d].WorkspacePath = %q, want %q", i, r.WorkspacePath, paths[i])
		}
	}
}

func TestRunAll_EmptyWorkspaces(t *testing.T) {
	results := runner.RunAll([]string{}, runner.Options{})
	if results == nil {
		// nil is acceptable but len must be 0
		return
	}
	if len(results) != 0 {
		t.Errorf("RunAll() with empty paths returned %d results, want 0", len(results))
	}
}

// buildFakeTerraform compiles a fake terraform binary from Go source and returns its path.
func buildFakeTerraform(t *testing.T, goSrc string) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not in PATH, skipping fake binary test")
	}
	dir := t.TempDir()
	srcFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcFile, []byte(goSrc), 0644); err != nil {
		t.Fatalf("buildFakeTerraform write: %v", err)
	}
	binPath := filepath.Join(dir, "fake-terraform")
	cmd := exec.Command("go", "build", "-o", binPath, srcFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("buildFakeTerraform compile: %v\n%s", err, out)
	}
	return binPath
}
