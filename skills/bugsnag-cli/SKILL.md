---
name: bugsnag-cli
description: Use the bugsnag-cli command-line tool to interact with the Bugsnag Data Access API. Use this skill whenever the user mentions Bugsnag errors, wants to investigate production bugs tracked in Bugsnag, check project health or stability metrics, list organizations/projects/releases, configure their Bugsnag CLI access, or needs help with any bugsnag-cli command syntax. Also triggers when the user asks about error trends, crash-free rates, or wants to comment on Bugsnag errors from the terminal.
---

# bugsnag-cli

A fast, agent-friendly CLI for the Bugsnag Data Access API. Outputs JSON by default (ideal for agents and scripts), supports `--format table` for human reading.

## When to read which reference

Pick the reference file that matches the user's intent:

| User intent | Reference file |
|-------------|---------------|
| Install, configure, authenticate, first steps | `references/quickstart.md` |
| Investigate a specific error (details, events, trends, comments) | `references/investigate-error.md` |
| Assess overall project health (error counts, trends, stability, releases) | `references/project-health.md` |
| Look up exact command syntax, flags, or options | `references/command-reference.md` |

If unsure, start with `references/command-reference.md` — it covers every command.

## Quick overview

### Authentication

Three methods (checked in this priority order):

1. Flag: `--api-token TOKEN` or `-t TOKEN`
2. Env var: `BUGSNAG_API_TOKEN`
3. Config file: `~/.bugsnag-cli.yaml` (created via `bugsnag configure --api-token TOKEN`)

### Available commands

| Command | Description |
|---------|-------------|
| `bugsnag configure` | Save auth token and defaults to config file |
| `bugsnag organizations list` | List organizations |
| `bugsnag projects list/get` | List or get project details |
| `bugsnag errors list/get` | List errors with filters, or get error details |
| `bugsnag events list/get` | List event occurrences, or get event details |
| `bugsnag trends project/error` | View error trends over time |
| `bugsnag collaborators list` | List organization collaborators |
| `bugsnag comments list/create` | List or add comments on errors |
| `bugsnag releases list` | List project releases |
| `bugsnag stability trend` | View crash-free session rate |
| `bugsnag version` | Print CLI version |

### Global flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--api-token` | `-t` | — | Bugsnag API token |
| `--format` | `-f` | `json` | Output format: json or table |
| `--per-page` | — | `30` | Results per page (1-100) |
| `--all-pages` | `-a` | `false` | Fetch all pages |
| `--base-url` | — | `https://api.bugsnag.com` | API base URL |
| `--config` | — | `~/.bugsnag-cli.yaml` | Config file path |

### Output format

JSON lists return: `{"data": [...], "total_count": N, "has_more": bool}`

Single items return the object directly. Errors go to stderr as `{"error": "..."}`.

### Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error (missing token/flag) |
| 3 | API error (401, 403, 404, 500) |
| 4 | Network error (timeout, DNS, connection refused) |

### Common workflows

**First setup:**
```bash
bugsnag configure --api-token TOKEN
bugsnag organizations list
bugsnag projects list --org-id ORG_ID
```

**Investigate an error:** read `references/investigate-error.md`

**Project health check:** read `references/project-health.md`

**Need exact flag syntax:** read `references/command-reference.md`
