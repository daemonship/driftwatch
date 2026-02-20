package report_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/daemonship/driftwatch/internal/report"
)

func noDriftResults() []report.ScanResult {
	return []report.ScanResult{
		{WorkspacePath: "./infra/staging", ResourceChanges: nil},
		{WorkspacePath: "./infra/production", ResourceChanges: nil},
	}
}

func driftResults() []report.ScanResult {
	return []report.ScanResult{
		{
			WorkspacePath: "./infra/staging",
			ResourceChanges: []report.ResourceChange{
				{
					Address: "aws_instance.web",
					Action:  "update",
					Attributes: map[string]report.AttributeChange{
						"ami": {Before: "ami-old", After: "ami-new"},
					},
				},
			},
		},
		{WorkspacePath: "./infra/production", ResourceChanges: nil},
	}
}

func errorResults() []report.ScanResult {
	return []report.ScanResult{
		{WorkspacePath: "./infra/staging", Err: errors.New("credentials not configured")},
	}
}

func TestExitCode_NoDrift(t *testing.T) {
	code := report.ExitCode(noDriftResults())
	if code != 0 {
		t.Errorf("ExitCode() = %d, want 0 for no drift", code)
	}
}

func TestExitCode_DriftDetected(t *testing.T) {
	code := report.ExitCode(driftResults())
	if code != 1 {
		t.Errorf("ExitCode() = %d, want 1 for drift detected", code)
	}
}

func TestExitCode_ScanError(t *testing.T) {
	code := report.ExitCode(errorResults())
	if code != 2 {
		t.Errorf("ExitCode() = %d, want 2 for scan error", code)
	}
}

func TestExitCode_DriftAndError(t *testing.T) {
	// When both drift and error exist, exit code should be 2 (error takes precedence).
	mixed := append(driftResults(), errorResults()...)
	code := report.ExitCode(mixed)
	if code != 2 {
		t.Errorf("ExitCode() = %d, want 2 when both drift and error present", code)
	}
}

func TestSummarize_NoDrift(t *testing.T) {
	summary := report.Summarize(noDriftResults())
	if summary.WorkspacesScanned != 2 {
		t.Errorf("WorkspacesScanned = %d, want 2", summary.WorkspacesScanned)
	}
	if summary.WorkspacesWithDrift != 0 {
		t.Errorf("WorkspacesWithDrift = %d, want 0", summary.WorkspacesWithDrift)
	}
	if summary.TotalDriftedResources != 0 {
		t.Errorf("TotalDriftedResources = %d, want 0", summary.TotalDriftedResources)
	}
	if summary.ScanErrors != 0 {
		t.Errorf("ScanErrors = %d, want 0", summary.ScanErrors)
	}
}

func TestSummarize_WithDrift(t *testing.T) {
	summary := report.Summarize(driftResults())
	if summary.WorkspacesScanned != 2 {
		t.Errorf("WorkspacesScanned = %d, want 2", summary.WorkspacesScanned)
	}
	if summary.WorkspacesWithDrift != 1 {
		t.Errorf("WorkspacesWithDrift = %d, want 1", summary.WorkspacesWithDrift)
	}
	if summary.TotalDriftedResources != 1 {
		t.Errorf("TotalDriftedResources = %d, want 1", summary.TotalDriftedResources)
	}
}

func TestSummarize_WithError(t *testing.T) {
	summary := report.Summarize(errorResults())
	if summary.ScanErrors != 1 {
		t.Errorf("ScanErrors = %d, want 1", summary.ScanErrors)
	}
}

func TestPrint_NoDriftOutput(t *testing.T) {
	var buf bytes.Buffer
	report.Print(&buf, noDriftResults())
	output := buf.String()
	if output == "" {
		t.Error("Print() produced no output, want at least a summary line")
	}
	// Should mention no drift or 0 drifted resources.
	if !strings.Contains(output, "0") && !strings.Contains(strings.ToLower(output), "no drift") {
		t.Errorf("Print() output %q does not mention zero drift", output)
	}
}

func TestPrint_ShowsWorkspacePath(t *testing.T) {
	var buf bytes.Buffer
	report.Print(&buf, driftResults())
	output := buf.String()
	if !strings.Contains(output, "./infra/staging") {
		t.Errorf("Print() output does not contain workspace path './infra/staging':\n%s", output)
	}
}

func TestPrint_ShowsResourceAddress(t *testing.T) {
	var buf bytes.Buffer
	report.Print(&buf, driftResults())
	output := buf.String()
	if !strings.Contains(output, "aws_instance.web") {
		t.Errorf("Print() output does not contain resource address 'aws_instance.web':\n%s", output)
	}
}

func TestPrint_ShowsBeforeAfterValues(t *testing.T) {
	var buf bytes.Buffer
	report.Print(&buf, driftResults())
	output := buf.String()
	if !strings.Contains(output, "ami-old") {
		t.Errorf("Print() output does not contain before value 'ami-old':\n%s", output)
	}
	if !strings.Contains(output, "ami-new") {
		t.Errorf("Print() output does not contain after value 'ami-new':\n%s", output)
	}
}

func TestPrint_ShowsSummaryCounts(t *testing.T) {
	var buf bytes.Buffer
	report.Print(&buf, driftResults())
	output := buf.String()
	// Output should include summary counts.
	if !strings.Contains(output, "2") {
		t.Errorf("Print() output should contain workspace count '2':\n%s", output)
	}
}

func TestPrint_ErrorWorkspace(t *testing.T) {
	var buf bytes.Buffer
	report.Print(&buf, errorResults())
	output := buf.String()
	if !strings.Contains(output, "./infra/staging") {
		t.Errorf("Print() output does not contain errored workspace path:\n%s", output)
	}
	if !strings.Contains(strings.ToLower(output), "error") {
		t.Errorf("Print() output does not mention error for errored workspace:\n%s", output)
	}
}
