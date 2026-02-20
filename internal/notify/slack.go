// Package notify sends drift notifications to external services.
package notify

import (
	"io"

	"github.com/daemonship/driftwatch/internal/report"
)

// SlackNotifier sends drift summaries to a Slack incoming webhook.
type SlackNotifier struct {
	// WebhookURL is the Slack incoming webhook URL.
	WebhookURL string
	// ErrOut is where HTTP errors are written (default: os.Stderr).
	ErrOut io.Writer
}

// Notify posts a drift summary to the Slack webhook if drift was detected.
// Silent (no POST) when results contain no drift.
// HTTP errors are written to ErrOut but do not return an error.
func (n *SlackNotifier) Notify(results []report.ScanResult) error {
	// TODO: implement in Task 5
	return nil
}

// WebhookFromEnv returns the Slack webhook URL from the DRIFTWATCH_SLACK_WEBHOOK
// environment variable, or empty string if not set.
func WebhookFromEnv() string {
	// TODO: implement in Task 5
	return ""
}
