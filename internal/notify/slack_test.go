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
