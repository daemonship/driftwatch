// Package report formats drift scan results for human-readable output.
package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/daemonship/driftwatch/internal/parser"
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
	WorkspacesScanned    int
	WorkspacesWithDrift  int
	TotalDriftedResources int
	ScanErrors           int
}

// ExitCode returns the appropriate process exit code for the scan results:
//
//	0 — no drift detected
//	1 — drift detected
//	2 — scan error occurred
func ExitCode(results []ScanResult) int {
	hasError := false
	hasDrift := false

	for _, r := range results {
		if r.Err != nil {
			hasError = true
			break
		}
		if len(r.ResourceChanges) > 0 {
			hasDrift = true
		}
	}

	if hasError {
		return 2
	}
	if hasDrift {
		return 1
	}
	return 0
}

// Print writes a human-readable drift report to w.
// Format: workspace → resource → attribute → before/after table.
func Print(w io.Writer, results []ScanResult) {
	summary := Summarize(results)

	// Print summary header
	fmt.Fprintf(w, "Drift Scan Summary\n")
	fmt.Fprintf(w, "===================\n")
	fmt.Fprintf(w, "Workspaces scanned: %d\n", summary.WorkspacesScanned)
	fmt.Fprintf(w, "Workspaces with drift: %d\n", summary.WorkspacesWithDrift)
	fmt.Fprintf(w, "Total drifted resources: %d\n", summary.TotalDriftedResources)
	fmt.Fprintf(w, "Scan errors: %d\n", summary.ScanErrors)
	fmt.Fprintln(w)

	// Print detailed results per workspace
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "ERROR: %s\n", r.WorkspacePath)
			fmt.Fprintf(w, "  %v\n", r.Err)
			continue
		}

		fmt.Fprintf(w, "Workspace: %s\n", r.WorkspacePath)

		if len(r.ResourceChanges) == 0 {
			fmt.Fprintf(w, "  No drift detected\n")
		} else {
			for _, rc := range r.ResourceChanges {
				fmt.Fprintf(w, "  Resource: %s (action: %s)\n", rc.Address, rc.Action)
				for attr, change := range rc.Attributes {
					beforeStr := formatValue(change.Before)
					afterStr := formatValue(change.After)
					fmt.Fprintf(w, "    %s:\n      before: %s\n      after:  %s\n", attr, beforeStr, afterStr)
				}
			}
		}
		fmt.Fprintln(w)
	}
}

// formatValue converts an interface{} value to a string for display.
func formatValue(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case []interface{}:
		parts := make([]string, len(val))
		for i, p := range val {
			parts[i] = formatValue(p)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		parts := make([]string, 0, len(val))
		for k, v := range val {
			parts = append(parts, fmt.Sprintf("%s=%s", k, formatValue(v)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Summarize computes aggregate statistics from scan results.
func Summarize(results []ScanResult) Summary {
	summary := Summary{
		WorkspacesScanned: len(results),
	}

	for _, r := range results {
		if r.Err != nil {
			summary.ScanErrors++
		} else if len(r.ResourceChanges) > 0 {
			summary.WorkspacesWithDrift++
			summary.TotalDriftedResources += len(r.ResourceChanges)
		}
	}

	return summary
}

// WorkspaceResultsFromRunnerResults converts runner results (with raw plan JSON)
// into report ScanResults. Requires parsing each plan.
func WorkspaceResultsFromRunnerResults(runnerResults []runner.Result) ([]ScanResult, error) {
	results := make([]ScanResult, 0, len(runnerResults))

	for _, r := range runnerResults {
		sr := ScanResult{
			WorkspacePath: r.WorkspacePath,
		}

		if r.Err != nil {
			sr.Err = r.Err
			results = append(results, sr)
			continue
		}

		// Parse the JSON plan output
		plan, err := parser.Parse(r.PlanOutput)
		if err != nil {
			sr.Err = fmt.Errorf("parsing plan JSON: %w", err)
			results = append(results, sr)
			continue
		}

		// Convert parser.ResourceChange to report.ResourceChange
		sr.ResourceChanges = make([]ResourceChange, 0, len(plan.ResourceChanges))
		for _, rc := range plan.ResourceChanges {
			attrs := make(map[string]AttributeChange)
			for attr, change := range rc.AttributeChanges {
				attrs[attr] = AttributeChange{
					Before: formatValue(change.Before),
					After:  formatValue(change.After),
				}
			}
			sr.ResourceChanges = append(sr.ResourceChanges, ResourceChange{
				Address:    rc.Address,
				Action:     string(rc.Action),
				Attributes: attrs,
			})
		}

		results = append(results, sr)
	}

	return results, nil
}
