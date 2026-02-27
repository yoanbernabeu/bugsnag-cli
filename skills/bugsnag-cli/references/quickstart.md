# Bugsnag CLI — Quick Start

## Prerequisites

The `bugsnag` binary must be available in PATH. Installation options:

```bash
# Via Go
go install github.com/yoanbernabeu/bugsnag-cli@latest

# Or one-liner install script
curl -fsSL https://raw.githubusercontent.com/yoanbernabeu/bugsnag-cli/main/install.sh | sh
```

## Step 1: Get an API Token

1. Log in to https://app.bugsnag.com
2. Go to **Settings > My account > Auth tokens**
3. Generate a new token (shown only once — store it safely)

## Step 2: Configure Authentication

Three ways to authenticate (checked in this priority order):

| Priority | Method | How |
|----------|--------|-----|
| 1 | Flag | `--api-token TOKEN` or `-t TOKEN` on every command |
| 2 | Env var | `export BUGSNAG_API_TOKEN=TOKEN` |
| 3 | Config file | `bugsnag configure --api-token TOKEN` (saves to ~/.bugsnag-cli.yaml) |

The recommended approach for persistent use:

```bash
bugsnag configure --api-token YOUR_TOKEN
```

Optional: set defaults for format, pagination, and base URL:

```bash
bugsnag configure --api-token YOUR_TOKEN \
  --default-format table \
  --default-per-page 50 \
  --default-base-url https://api.bugsnag.com
```

The config file is written with mode 0600 (readable only by the owner).

## Step 3: Verify Setup

```bash
# List organizations — confirms auth works
bugsnag organizations list

# List projects for an org
bugsnag projects list --org-id ORG_ID
```

## Step 4: Explore Your Data

```bash
# List errors for a project (JSON)
bugsnag errors list --project-id PROJECT_ID

# Human-readable output
bugsnag errors list --project-id PROJECT_ID --format table

# Fetch all pages at once
bugsnag errors list --project-id PROJECT_ID --all-pages

# Filter to critical open errors
bugsnag errors list --project-id PROJECT_ID --status open --severity error
```

## Typical First Session

1. `bugsnag configure --api-token TOKEN` — save credentials
2. `bugsnag organizations list` — find your org ID
3. `bugsnag projects list --org-id ORG_ID` — find your project ID
4. `bugsnag errors list --project-id PROJECT_ID --format table` — see recent errors
5. `bugsnag errors list --project-id PROJECT_ID --status open --severity error` — filter to critical open errors

## Environment Variables

| Env Var | Maps to |
|---------|---------|
| `BUGSNAG_API_TOKEN` | `--api-token` |
| `BUGSNAG_FORMAT` | `--format` |
| `BUGSNAG_PER_PAGE` | `--per-page` |
| `BUGSNAG_BASE_URL` | `--base-url` |
