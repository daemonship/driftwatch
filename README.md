# driftwatch

> Terraform state drift is a silent killer for small infra teams. Manually checking is tedious, and existing enterprise tools are overkill and expensive.

`driftwatch` is a single static binary that scans your Terraform workspaces for drift by running `terraform plan` and surfacing any resource changes in a clear, actionable report.

**Requires Terraform >= 1.0.0**

## Feedback & Ideas

> **This project is being built in public and we want to hear from you.**
> Found a bug? Have a feature idea? Something feel wrong or missing?
> **[Open an issue](../../issues)** — every piece of feedback directly shapes what gets built next.

## Status

> 🚧 In active development — not yet production ready

| Feature | Status | Notes |
|---------|--------|-------|
| Project scaffold & CI | ✅ Complete | Go module, cobra CLI, goreleaser, GitHub Actions |
| Config loader & workspace runner | ✅ Complete | Config parsing, workspace runner with terraform plan execution |
| Terraform plan JSON parser | ✅ Complete | Parses resource_changes, diffs attributes, handles create/update/delete/replace |
| Drift summary report & exit codes | 🚧 In Progress | |
| Slack webhook notification | 📋 Planned | |

## What It Solves

Small teams managing cloud infra with Terraform often discover drift only when things break. `driftwatch` gives you a fast, scriptable way to check all your workspaces at once — in CI, in a cron job, or on demand.

## Who It's For

DevOps engineers or full-stack developers at startups and small teams managing cloud infra with Terraform.

## Usage (coming soon)

```bash
# Install (once released)
curl -L https://github.com/daemonship/driftwatch/releases/latest/download/driftwatch_linux_amd64.tar.gz | tar xz
sudo mv driftwatch /usr/local/bin/

# Configure your workspaces
cp driftwatch.yml.example driftwatch.yml
# edit driftwatch.yml to add your workspace paths

# Scan for drift
driftwatch scan

# Exit codes:
#   0 — no drift
#   1 — drift detected
#   2 — scan error
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
