<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="MIT License">
  <a href="https://github.com/yoanbernabeu/bugsnag-cli/actions/workflows/ci.yml"><img src="https://github.com/yoanbernabeu/bugsnag-cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/yoanbernabeu/bugsnag-cli/releases"><img src="https://img.shields.io/github/v/release/yoanbernabeu/bugsnag-cli" alt="Release"></a>
</p>

# bugsnag-cli

> A fast, agent-friendly CLI for the [Bugsnag Data Access API](https://bugsnagapiv2.docs.apiary.io/).

**bugsnag-cli** gives you full read access to your Bugsnag data from the terminal. It outputs structured **JSON by default** (perfect for code agents, scripts, and CI pipelines) and supports a human-readable `--format table` mode.

---

## Features

- **All Bugsnag resources** — organizations, projects, errors, events, trends, collaborators, comments, releases, stability
- **JSON-first** — structured envelope `{ "data": [...], "total_count": N, "has_more": bool }` for lists
- **Table output** — `--format table` for quick human inspection
- **Auto-pagination** — `--all-pages` fetches every page in one go
- **Agent-optimized** — deterministic exit codes, errors on stderr, no interactive prompts, no noisy help on failure
- **Flexible auth** — flag, env var, or config file (priority order)
- **Cross-platform** — Linux, macOS, Windows (amd64 & arm64)

---

## Installation

### One-liner (Linux / macOS / Windows)

```bash
curl -fsSL https://raw.githubusercontent.com/yoanbernabeu/bugsnag-cli/main/install.sh | sh
```

> On Windows (PowerShell):
> ```powershell
> irm https://raw.githubusercontent.com/yoanbernabeu/bugsnag-cli/main/install.ps1 | iex
> ```

### From releases

Download the latest binary for your platform from the [Releases page](https://github.com/yoanbernabeu/bugsnag-cli/releases).

### With Go

```bash
go install github.com/yoanbernabeu/bugsnag-cli@latest
```

### From source

```bash
git clone https://github.com/yoanbernabeu/bugsnag-cli.git
cd bugsnag-cli
go build -o bugsnag .
```

---

## Agent Skills

Install the bugsnag-cli skill to let your AI coding agent use the CLI. Compatible with [Claude Code](https://claude.com/claude-code), [Cursor](https://cursor.com), [Codex](https://openai.com/codex), [GitHub Copilot](https://github.com/features/copilot), [Windsurf](https://windsurf.com), and [more](https://github.com/rohitg00/skillkit):

```bash
npx skills add https://github.com/yoanbernabeu/bugsnag-cli --skill bugsnag-cli
```

---

## Getting your API Token

1. Log in to [Bugsnag](https://app.bugsnag.com)
2. Go to **Settings > My account > Auth tokens**
   Direct link: `https://app.bugsnag.com/settings/<YOUR_ORG>/my-account/auth-tokens`
3. Click **Generate new token**, give it a description, and copy the token

> **Note:** The token is shown only once. Store it safely.

---

## Quick Start

```bash
# 1. Save your token (written to ~/.bugsnag-cli.yaml with 0600 permissions)
bugsnag configure --api-token YOUR_TOKEN

# 2. List your organizations
bugsnag organizations list

# 3. List errors for a project
bugsnag errors list --project-id PROJECT_ID

# 4. Same thing, but as a table
bugsnag errors list --project-id PROJECT_ID --format table

# 5. Fetch ALL errors across every page
bugsnag errors list --project-id PROJECT_ID --all-pages
```

---

## Configuration

Authentication is resolved in this order:

| Priority | Method | Example |
|----------|--------|---------|
| 1 | Flag | `--api-token TOKEN` or `-t TOKEN` |
| 2 | Env var | `export BUGSNAG_API_TOKEN=TOKEN` |
| 3 | Config file | `~/.bugsnag-cli.yaml` |

### Persistent config

```bash
bugsnag configure --api-token YOUR_TOKEN
```

This writes `~/.bugsnag-cli.yaml` with mode `0600`. Optional flags:

```bash
bugsnag configure --api-token TOKEN \
  --default-format table \
  --default-per-page 50 \
  --default-base-url https://api.bugsnag.com
```

### Config file format

```yaml
api_token: your-token-here
format: json        # or "table"
per_page: 30
base_url: https://api.bugsnag.com
```

---

## Global Flags

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--api-token`, `-t` | `BUGSNAG_API_TOKEN` | *required* | Bugsnag API token |
| `--format`, `-f` | `BUGSNAG_FORMAT` | `json` | Output format: `json` or `table` |
| `--per-page` | `BUGSNAG_PER_PAGE` | `30` | Results per page (1–100) |
| `--all-pages`, `-a` | — | `false` | Fetch all pages automatically |
| `--base-url` | `BUGSNAG_BASE_URL` | `https://api.bugsnag.com` | API base URL |
| `--config` | — | `~/.bugsnag-cli.yaml` | Path to config file |

---

## Commands

### Organizations

```bash
bugsnag organizations list
```

### Projects

```bash
bugsnag projects list --org-id ORG_ID
bugsnag projects get  --project-id PROJECT_ID
```

### Errors

```bash
bugsnag errors list --project-id ID
bugsnag errors list --project-id ID --status open --severity error
bugsnag errors list --project-id ID --sort last_seen --direction desc
bugsnag errors get  --project-id ID --error-id ERROR_ID
```

### Events

```bash
bugsnag events list --project-id ID
bugsnag events list --project-id ID --error-id ERROR_ID
bugsnag events get  --project-id ID --event-id EVENT_ID
```

### Trends

```bash
bugsnag trends project --project-id ID
bugsnag trends project --project-id ID --resolution 1d --buckets-count 14
bugsnag trends error   --project-id ID --error-id ERROR_ID
```

### Collaborators

```bash
bugsnag collaborators list --org-id ORG_ID
```

### Comments

```bash
bugsnag comments list   --project-id ID --error-id ERROR_ID
bugsnag comments create --project-id ID --error-id ERROR_ID --message "Fixed in v2.1"
```

### Releases

```bash
bugsnag releases list --project-id ID
```

### Stability

```bash
bugsnag stability trend --project-id ID
bugsnag stability trend --project-id ID --release-stage production
```

### Utility

```bash
bugsnag version
bugsnag configure --api-token TOKEN
```

---

## Output Format

### JSON (default)

**Lists** return a structured envelope:

```json
{
  "data": [
    { "id": "abc123", "class": "NoMethodError", "message": "undefined method" }
  ],
  "total_count": 1,
  "has_more": false
}
```

**Single items** return the object directly:

```json
{
  "id": "abc123",
  "class": "NoMethodError",
  "message": "undefined method 'foo' for nil:NilClass"
}
```

**Errors** (on stderr):

```json
{"error": "API error (401): Bad Credentials"}
```

### Table

```
ID        CLASS           MESSAGE                  STATUS  SEVERITY  EVENTS
abc123    NoMethodError   undefined method 'foo'   open    error     42
def456    TypeError       nil is not a string      open    warning   7
```

---

## Pagination

By default, one page of results is returned. Use `--all-pages` to collect everything:

```bash
bugsnag errors list --project-id ID --all-pages
```

The CLI follows Bugsnag's `Link` header pagination automatically.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Configuration error (missing token, missing flag) |
| `3` | API error (401, 403, 404, 500, etc.) |
| `4` | Network error (timeout, DNS, connection refused) |

---

## For Code Agents

This CLI is built with automation in mind:

- **JSON by default** — parse `.data` for list results
- **Structured errors on stderr** — `{"error": "..."}` in JSON mode
- **Deterministic exit codes** — branch on `$?` to classify failures
- **No interactive prompts** — safe for unattended execution
- **Silent usage** — no help text dumped on errors

Example in a script:

```bash
errors=$(bugsnag errors list --project-id "$PID" --all-pages 2>/dev/null)
count=$(echo "$errors" | jq '.total_count')
echo "Found $count errors"
```

---

## Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Lint
golangci-lint run

# Build
go build -o bugsnag .
```

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a PR.

## Security

To report a vulnerability, see [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) &copy; Yoan Bernabeu
