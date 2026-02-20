// Package notify sends drift notifications to external services.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/daemonship/driftwatch/internal/report"
)

// SlackNotifier sends drift summaries to a Slack incoming webhook.
type SlackNotifier struct {
	// WebhookURL is the Slack incoming webhook URL.
	WebhookURL string
	// ErrOut is where HTTP errors are written (default: os.Stderr).
	ErrOut io.Writer
}

// slackMessage represents the JSON payload sent to a Slack webhook.
type slackMessage struct {
	Text        string            `json:"text"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

// slackAttachment represents a Slack message attachment.
type slackAttachment struct {
	Color     string   `json:"color"`
	Title     string   `json:"title"`
	Text      string   `json:"text"`
	MrkdwnIn  []string `json:"mrkdwn_in,omitempty"`
	Timestamp int64    `json:"ts"`
}

// Notify posts a drift summary to the Slack webhook if drift was detected.
// Silent (no POST) when results contain no drift.
// HTTP errors are written to ErrOut but do not return an error.
func (n *SlackNotifier) Notify(results []report.ScanResult) error {
	// Check if there's any drift to report
	summary := report.Summarize(results)
	if summary.WorkspacesWithDrift == 0 {
		// Silent on clean scans
		return nil
	}

	// Default ErrOut to os.Stderr if not set
	errOut := n.ErrOut
	if errOut == nil {
		errOut = os.Stderr
	}

	// Build the Slack message
	msg := n.buildSlackMessage(results, summary)

	// Marshal to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		fmt.Fprintf(errOut, "Error marshaling Slack message: %v\n", err)
		return nil // Don't return error, just log
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", n.WebhookURL, bytes.NewReader(jsonData))
	if err != nil {
		fmt.Fprintf(errOut, "Error creating Slack request: %v\n", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(errOut, "Error sending Slack notification: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	// Check for HTTP errors (but don't fail the scan)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(errOut, "Slack webhook returned HTTP %d\n", resp.StatusCode)
	}

	return nil
}

// buildSlackMessage constructs a Slack message from scan results.
func (n *SlackNotifier) buildSlackMessage(results []report.ScanResult, summary report.Summary) slackMessage {
	// Build a concise summary message
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("*Drift Detected in %d Workspace(s)*\n\n", summary.WorkspacesWithDrift))
	buf.WriteString(fmt.Sprintf("Total drifted resources: %d\n", summary.TotalDriftedResources))

	// List affected workspaces
	var affectedWorkspaces []string
	for _, r := range results {
		if len(r.ResourceChanges) > 0 {
			affectedWorkspaces = append(affectedWorkspaces, r.WorkspacePath)
		}
	}

	if len(affectedWorkspaces) > 0 {
		buf.WriteString("\n*Affected Workspaces:*\n")
		for _, ws := range affectedWorkspaces {
			buf.WriteString(fmt.Sprintf("• %s\n", ws))
		}
	}

	// Add a summary of changes by workspace
	buf.WriteString("\n*Changes Summary:*\n")
	for _, r := range results {
		if len(r.ResourceChanges) == 0 {
			continue
		}

		buf.WriteString(fmt.Sprintf("\n%s:\n", r.WorkspacePath))
		for _, rc := range r.ResourceChanges {
			buf.WriteString(fmt.Sprintf("  • `%s` (%s)\n", rc.Address, rc.Action))
		}
	}

	// Create attachment with color based on severity
	color := "warning" // yellow/orange for drift
	if summary.WorkspacesWithDrift > 1 {
		color = "danger" // red for multiple workspaces
	}

	return slackMessage{
		Text: "🚨 Terraform Drift Detected",
		Attachments: []slackAttachment{
			{
				Color:     color,
				Title:     "Drift Summary",
				Text:      buf.String(),
				MrkdwnIn:  []string{"text"},
				Timestamp: time.Now().Unix(),
			},
		},
	}
}

// WebhookFromEnv returns the Slack webhook URL from the DRIFTWATCH_SLACK_WEBHOOK
// environment variable, or empty string if not set.
func WebhookFromEnv() string {
	return os.Getenv("DRIFTWATCH_SLACK_WEBHOOK")
}
