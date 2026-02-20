package notify_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/daemonship/driftwatch/internal/notify"
	"github.com/daemonship/driftwatch/internal/report"
)

func noDriftResults() []report.ScanResult {
	return []report.ScanResult{
		{WorkspacePath: "./infra/staging"},
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
	}
}

func TestNotify_SilentOnNoDrift(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &notify.SlackNotifier{WebhookURL: srv.URL}
	if err := n.Notify(noDriftResults()); err != nil {
		t.Errorf("Notify() error = %v, want nil", err)
	}
	if called {
		t.Error("Notify() made HTTP request for no-drift results, want silent")
	}
}

func TestNotify_PostsOnDrift(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &notify.SlackNotifier{WebhookURL: srv.URL}
	if err := n.Notify(driftResults()); err != nil {
		t.Errorf("Notify() error = %v, want nil", err)
	}
	if len(body) == 0 {
		t.Error("Notify() sent empty body to Slack webhook")
	}
}

func TestNotify_BodyContainsWorkspace(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &notify.SlackNotifier{WebhookURL: srv.URL}
	if err := n.Notify(driftResults()); err != nil {
		t.Errorf("Notify() error = %v, want nil", err)
	}
	if !strings.Contains(string(body), "staging") {
		t.Errorf("Notify() body %q does not mention workspace 'staging'", body)
	}
}

func TestNotify_BodyContainsResourceAddress(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &notify.SlackNotifier{WebhookURL: srv.URL}
	if err := n.Notify(driftResults()); err != nil {
		t.Errorf("Notify() error = %v, want nil", err)
	}
	if !strings.Contains(string(body), "aws_instance.web") {
		t.Errorf("Notify() body %q does not contain resource address", body)
	}
}

func TestNotify_HTTPErrorDoesNotReturnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var errBuf strings.Builder
	n := &notify.SlackNotifier{
		WebhookURL: srv.URL,
		ErrOut:     &errBuf,
	}
	// HTTP 500 should NOT return an error — just log to stderr.
	if err := n.Notify(driftResults()); err != nil {
		t.Errorf("Notify() error = %v, want nil even on HTTP error", err)
	}
	if errBuf.Len() == 0 {
		t.Error("Notify() did not write HTTP error to ErrOut")
	}
}

func TestNotify_InvalidURLDoesNotPanic(t *testing.T) {
	var errBuf strings.Builder
	n := &notify.SlackNotifier{
		WebhookURL: "http://127.0.0.1:1", // nothing listening
		ErrOut:     &errBuf,
	}
	// Connection error should not return error — just log.
	if err := n.Notify(driftResults()); err != nil {
		t.Errorf("Notify() error = %v, want nil on connection error", err)
	}
}

func TestWebhookFromEnv_NotSet(t *testing.T) {
	os.Unsetenv("DRIFTWATCH_SLACK_WEBHOOK")
	url := notify.WebhookFromEnv()
	if url != "" {
		t.Errorf("WebhookFromEnv() = %q, want empty string when env not set", url)
	}
}

func TestWebhookFromEnv_Set(t *testing.T) {
	expected := "https://hooks.slack.com/services/test/url"
	os.Setenv("DRIFTWATCH_SLACK_WEBHOOK", expected)
	t.Cleanup(func() { os.Unsetenv("DRIFTWATCH_SLACK_WEBHOOK") })

	url := notify.WebhookFromEnv()
	if url != expected {
		t.Errorf("WebhookFromEnv() = %q, want %q", url, expected)
	}
}

func TestNotify_UsesJSONContentType(t *testing.T) {
	var contentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &notify.SlackNotifier{WebhookURL: srv.URL}
	if err := n.Notify(driftResults()); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
}

// TestNotify_HTTP200IsSuccess ensures status 200 does not write to ErrOut.
// Kills CONDITIONALS_BOUNDARY mutant on `resp.StatusCode < 200`.
func TestNotify_HTTP200IsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200
	}))
	defer srv.Close()

	var errBuf strings.Builder
	n := &notify.SlackNotifier{WebhookURL: srv.URL, ErrOut: &errBuf}
	if err := n.Notify(driftResults()); err != nil {
		t.Errorf("Notify() error = %v, want nil", err)
	}
	if errBuf.Len() > 0 {
		t.Errorf("Notify() wrote to ErrOut on HTTP 200 success: %q", errBuf.String())
	}
}

// TestNotify_HTTP300IsError ensures status 300 (redirect boundary) writes to ErrOut.
// Kills CONDITIONALS_BOUNDARY mutant on `resp.StatusCode >= 300`.
func TestNotify_HTTP300IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 300 is outside 2xx — must log an error
		w.WriteHeader(300)
	}))
	defer srv.Close()

	var errBuf strings.Builder
	n := &notify.SlackNotifier{WebhookURL: srv.URL, ErrOut: &errBuf}
	if err := n.Notify(driftResults()); err != nil {
		t.Errorf("Notify() error = %v, want nil", err)
	}
	if errBuf.Len() == 0 {
		t.Error("Notify() did not write to ErrOut for HTTP 300 status")
	}
}

func multiWorkspaceDriftResults() []report.ScanResult {
	change := report.ResourceChange{
		Address: "aws_instance.web",
		Action:  "update",
		Attributes: map[string]report.AttributeChange{
			"ami": {Before: "ami-old", After: "ami-new"},
		},
	}
	return []report.ScanResult{
		{WorkspacePath: "./infra/staging", ResourceChanges: []report.ResourceChange{change}},
		{WorkspacePath: "./infra/production", ResourceChanges: []report.ResourceChange{change}},
	}
}

// TestNotify_SingleWorkspaceDriftColorIsWarning verifies color="warning" for 1 drifted workspace.
// Kills CONDITIONALS_BOUNDARY and CONDITIONALS_NEGATION on `summary.WorkspacesWithDrift > 1`.
func TestNotify_SingleWorkspaceDriftColorIsWarning(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &notify.SlackNotifier{WebhookURL: srv.URL}
	if err := n.Notify(driftResults()); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if !strings.Contains(string(body), `"warning"`) {
		t.Errorf("expected color=warning for single workspace drift, got body: %s", body)
	}
	if strings.Contains(string(body), `"danger"`) {
		t.Errorf("unexpected color=danger for single workspace drift, got body: %s", body)
	}
}

// TestNotify_MultiWorkspaceDriftColorIsDanger verifies color="danger" for 2+ drifted workspaces.
// Kills CONDITIONALS_BOUNDARY and CONDITIONALS_NEGATION on `summary.WorkspacesWithDrift > 1`.
func TestNotify_MultiWorkspaceDriftColorIsDanger(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &notify.SlackNotifier{WebhookURL: srv.URL}
	if err := n.Notify(multiWorkspaceDriftResults()); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if !strings.Contains(string(body), `"danger"`) {
		t.Errorf("expected color=danger for multi workspace drift, got body: %s", body)
	}
}

// TestNotify_MixedResultsOnlyListsDriftedWorkspaces verifies that workspaces without
// drift do not appear in the affected-workspace list.
// Kills CONDITIONALS_BOUNDARY and CONDITIONALS_NEGATION on `len(r.ResourceChanges) > 0`.
func TestNotify_MixedResultsOnlyListsDriftedWorkspaces(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	results := []report.ScanResult{
		{
			WorkspacePath: "./infra/staging",
			ResourceChanges: []report.ResourceChange{
				{Address: "aws_instance.web", Action: "update"},
			},
		},
		{
			WorkspacePath:   "./infra/clean",
			ResourceChanges: nil, // no drift
		},
	}

	n := &notify.SlackNotifier{WebhookURL: srv.URL}
	if err := n.Notify(results); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "staging") {
		t.Errorf("expected 'staging' (has drift) to appear in body, got: %s", bodyStr)
	}
	if strings.Contains(bodyStr, "clean") {
		t.Errorf("expected 'clean' (no drift) to NOT appear in affected list, got: %s", bodyStr)
	}
}
