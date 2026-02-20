# driftwatch

> Terraform state drift is a silent killer for small infra teams. Manually checking is tedious, and existing enterprise tools are overkill and expensive.

`driftwatch` is a single static binary that scans your Terraform workspaces for drift by running `terraform plan` and surfacing any resource changes in a clear, actionable report.

**Requires Terraform >= 1.0.0**

## Feedback & Ideas

> **This project is being built in public and we want to hear from you.**
> Found a bug? Have a feature idea? Something feel wrong or missing?
> **[Open an issue](../../issues)** — every piece of feedback directly shapes what gets built next.

## Status

> 🚀 Ready for early adopters

| Feature | Status | Notes |
|---------|--------|-------|
| Project scaffold & CI | ✅ Complete | Go module, cobra CLI, goreleaser, GitHub Actions |
| Config loader & workspace runner | ✅ Complete | Config parsing, workspace runner with terraform plan execution |
| Terraform plan JSON parser | ✅ Complete | Parses resource_changes, diffs attributes, handles create/update/delete/replace |
| Drift summary report & exit codes | ✅ Complete | Exit codes 0/1/2, human-readable table output, summary counts |
| Slack webhook notification | ✅ Complete | Env var or config support, summary format (not full table), silent on clean scans |
| Ship-check & hardening | ✅ Complete | .gitignore, Go 1.24 (stdlib vulns fixed), goreleaser v2 format, 82% mutation score |

## What It Solves

Small teams managing cloud infra with Terraform often discover drift only when things break. `driftwatch` gives you a fast, scriptable way to check all your workspaces at once — in CI, in a cron job, or on demand.

## Who It's For

DevOps engineers or full-stack developers at startups and small teams managing cloud infra with Terraform.

## Installation

**Download a pre-built binary** (Linux/macOS/Windows, amd64/arm64):

```bash
# Linux amd64
curl -L https://github.com/daemonship/driftwatch/releases/latest/download/driftwatch_linux_amd64.tar.gz | tar xz
sudo mv driftwatch /usr/local/bin/

# macOS arm64 (Apple Silicon)
curl -L https://github.com/daemonship/driftwatch/releases/latest/download/driftwatch_darwin_arm64.tar.gz | tar xz
sudo mv driftwatch /usr/local/bin/
```

**Build from source** (requires Go 1.24+):

```bash
git clone https://github.com/daemonship/driftwatch.git
cd driftwatch
go build -o driftwatch .
```

## Usage

```bash
# 1. Copy the example config and add your workspace paths
cp driftwatch.yml.example driftwatch.yml

# 2. Scan all configured workspaces for drift
driftwatch scan

# Exit codes:
#   0 — no drift detected
#   1 — drift detected in one or more workspaces
#   2 — scan error (terraform not found, plan failed, etc.)
```

**Slack notifications** — set the webhook via env var (recommended) or config:

```bash
export DRIFTWATCH_SLACK_WEBHOOK=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
driftwatch scan
```

## Configuration

```yaml
# driftwatch.yml
workspaces:
  - ./infra/staging
  - ./infra/production

# Optional: Slack notifications on drift
# slack_webhook: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
# Or set DRIFTWATCH_SLACK_WEBHOOK env var

# Optional: use OpenTofu instead of Terraform
# binary: tofu
```

## Tech Stack

- **Go** — single static binary, no runtime deps
- **cobra** — CLI framework
- **gopkg.in/yaml.v3** — config parsing
- **goreleaser** — cross-platform release builds (linux/darwin, amd64/arm64)

---

*Built by [DaemonShip](https://github.com/daemonship) — autonomous venture studio*
