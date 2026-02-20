// Package report formats drift scan results for human-readable output.
package report

import (
	"io"

	"github.com/daemonship/driftwatch/internal/runner"
)

// ScanResult holds the parsed drift data for a single workspace.
type ScanResult struct {
	// WorkspacePath is the directory that was scanned.
	WorkspacePath string
	// ResourceChanges holds any drifted resources found.
	ResourceChanges []ResourceChange
	// Err is set if the workspace could not be scanned.
	Err error
}

// ResourceChange is a report-level resource change (for display).
type ResourceChange struct {
	Address    string
	Action     string
	Attributes map[string]AttributeChange
}

// AttributeChange is a report-level attribute change (for display).
type AttributeChange struct {
	Before string
	After  string
}

// Summary contains aggregate drift statistics.
type Summary struct {
	WorkspacesScanned  int
	WorkspacesWithDrift int
	TotalDriftedResources int
	ScanErrors         int
}

// ExitCode returns the appropriate process exit code for the scan results:
//
//	0 — no drift detected
//	1 — drift detected
//	2 — scan error occurred
func ExitCode(results []ScanResult) int {
	// TODO: implement in Task 4
	return 0
}

// Print writes a human-readable drift report to w.
// Format: workspace → resource → attribute → before/after table.
func Print(w io.Writer, results []ScanResult) {
	// TODO: implement in Task 4
}

// Summarize computes aggregate statistics from scan results.
func Summarize(results []ScanResult) Summary {
	// TODO: implement in Task 4
	return Summary{}
}

// WorkspaceResultsFromRunnerResults converts runner results (with raw plan JSON)
// into report ScanResults. Requires parsing each plan.
func WorkspaceResultsFromRunnerResults(runnerResults []runner.Result) ([]ScanResult, error) {
	// TODO: implement in Task 4
	return nil, nil
}
