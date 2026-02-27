# bugsnag-cli Command Reference

Complete reference for all commands. All flags verified against actual CLI help output (v0.1.0).

## Global Flags

| Flag | Short | Default | Env Var | Description |
|------|-------|---------|---------|-------------|
| `--api-token` | `-t` | — | `BUGSNAG_API_TOKEN` | Bugsnag API token |
| `--format` | `-f` | `json` | `BUGSNAG_FORMAT` | Output format: json or table |
| `--per-page` | — | `30` | `BUGSNAG_PER_PAGE` | Results per page (1-100) |
| `--all-pages` | `-a` | `false` | — | Fetch all pages |
| `--base-url` | — | `https://api.bugsnag.com` | `BUGSNAG_BASE_URL` | API base URL |
| `--config` | — | `~/.bugsnag-cli.yaml` | — | Config file path |

Auth priority: Flag > Env var > Config file.

---

## configure

Save configuration to ~/.bugsnag-cli.yaml (permissions 0600).

```bash
bugsnag configure --api-token TOKEN [--default-format FORMAT] [--default-per-page N] [--default-base-url URL]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--api-token`, `-t` | Yes | API token to save |
| `--default-format` | No | Default output format (json or table) |
| `--default-per-page` | No | Default results per page |
| `--default-base-url` | No | Default API base URL |

---

## organizations list

List organizations for the authenticated user.

```bash
bugsnag organizations list
```

No additional flags.

---

## projects list

```bash
bugsnag projects list --org-id ORG_ID
```

| Flag | Required | Description |
|------|----------|-------------|
| `--org-id` | Yes | Organization ID |

## projects get

```bash
bugsnag projects get --project-id PROJECT_ID
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |

---

## errors list

```bash
bugsnag errors list --project-id ID [--status STATUS] [--severity SEV] [--sort FIELD] [--direction DIR]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--status` | No | Filter: open, fixed, snoozed, ignored |
| `--severity` | No | Filter: info, warning, error |
| `--sort` | No | Sort by: created_at, last_seen, events, users, unsorted |
| `--direction` | No | Sort direction: asc, desc |

## errors get

```bash
bugsnag errors get --project-id ID --error-id ERROR_ID
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--error-id` | Yes | Error ID |

---

## events list

```bash
bugsnag events list --project-id ID [--error-id ERROR_ID]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--error-id` | No | Scope events to a specific error |

## events get

```bash
bugsnag events get --project-id ID --event-id EVENT_ID
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--event-id` | Yes | Event ID |

---

## trends project

```bash
bugsnag trends project --project-id ID [--resolution RES] [--buckets-count N]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--resolution` | No | Time resolution (e.g., 1h, 1d) |
| `--buckets-count` | No | Number of trend buckets |

## trends error

```bash
bugsnag trends error --project-id ID --error-id ERROR_ID
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--error-id` | Yes | Error ID |

**Important:** `trends error` does NOT accept `--resolution` or `--buckets-count`.

---

## collaborators list

```bash
bugsnag collaborators list --org-id ORG_ID
```

| Flag | Required | Description |
|------|----------|-------------|
| `--org-id` | Yes | Organization ID |

---

## comments list

```bash
bugsnag comments list --project-id ID --error-id ERROR_ID
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--error-id` | Yes | Error ID |

## comments create

```bash
bugsnag comments create --project-id ID --error-id ERROR_ID --message "Your comment"
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--error-id` | Yes | Error ID |
| `--message` | Yes | Comment text |

---

## releases list

```bash
bugsnag releases list --project-id ID
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |

---

## stability trend

```bash
bugsnag stability trend --project-id ID [--release-stage STAGE]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--project-id` | Yes | Project ID |
| `--release-stage` | No | Release stage (e.g., production, staging) |

---

## version

```bash
bugsnag version
```

---

## Output Format

### JSON (default)

Lists: `{"data": [...], "total_count": N, "has_more": bool}`

Single items: object directly (no envelope).

Errors on stderr: `{"error": "API error (401): Bad Credentials"}`

### Table

Use `--format table` for human-readable columnar output.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error (missing token, missing required flag) |
| 3 | API error (HTTP 401, 403, 404, 500, etc.) |
| 4 | Network error (timeout, DNS failure, connection refused) |
