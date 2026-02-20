package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daemonship/driftwatch/internal/config"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
workspaces:
  - ./infra/staging
  - ./infra/production
slack_webhook: https://hooks.slack.com/services/xxx/yyy/zzz
binary: terraform
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if len(cfg.Workspaces) != 2 {
		t.Errorf("Workspaces count = %d, want 2", len(cfg.Workspaces))
	}
	if cfg.Workspaces[0] != "./infra/staging" {
		t.Errorf("Workspaces[0] = %q, want %q", cfg.Workspaces[0], "./infra/staging")
	}
	if cfg.Workspaces[1] != "./infra/production" {
		t.Errorf("Workspaces[1] = %q, want %q", cfg.Workspaces[1], "./infra/production")
	}
	if cfg.SlackWebhook != "https://hooks.slack.com/services/xxx/yyy/zzz" {
		t.Errorf("SlackWebhook = %q, want webhook URL", cfg.SlackWebhook)
	}
	if cfg.Binary != "terraform" {
		t.Errorf("Binary = %q, want %q", cfg.Binary, "terraform")
	}
}

func TestLoad_MinimalConfig(t *testing.T) {
	content := `
workspaces:
  - ./infra
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if len(cfg.Workspaces) != 1 {
		t.Errorf("Workspaces count = %d, want 1", len(cfg.Workspaces))
	}
	if cfg.SlackWebhook != "" {
		t.Errorf("SlackWebhook = %q, want empty string", cfg.SlackWebhook)
	}
	if cfg.Binary != "" {
		t.Errorf("Binary = %q, want empty string", cfg.Binary)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/driftwatch.yml")
	if err == nil {
		t.Error("Load() error = nil, want error for missing file")
	}
}

func TestLoad_EmptyWorkspaceList(t *testing.T) {
	content := `
workspaces: []
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if len(cfg.Workspaces) != 0 {
		t.Errorf("Workspaces count = %d, want 0", len(cfg.Workspaces))
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := "this: is: not: valid: yaml: [[["
	path := writeTempConfig(t, content)
	_, err := config.Load(path)
	if err == nil {
		t.Error("Load() error = nil, want error for invalid YAML")
	}
}

func TestLoad_MissingWorkspacesKey(t *testing.T) {
	content := `
slack_webhook: https://hooks.slack.com/services/xxx
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil (missing workspaces is valid)", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if len(cfg.Workspaces) != 0 {
		t.Errorf("Workspaces count = %d, want 0", len(cfg.Workspaces))
	}
}

// writeTempConfig writes content to a temp file and returns its path.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "driftwatch.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return path
}
