# Bugsnag Project Health Assessment

Assess the overall health of a project — error counts, trends, stability, and release status.

## Prerequisites

- bugsnag-cli configured with a valid API token
- A project ID (run `bugsnag projects list --org-id ORG_ID` to find it)

## Health Assessment Workflow

### Step 1: Project Overview

```bash
bugsnag projects get --project-id PROJECT_ID
```

### Step 2: Error Summary

```bash
# Open errors sorted by impact (most events first)
bugsnag errors list --project-id PROJECT_ID --status open --sort events --direction desc

# Critical errors only
bugsnag errors list --project-id PROJECT_ID --status open --severity error

# All open errors across all pages
bugsnag errors list --project-id PROJECT_ID --status open --all-pages
```

Parse the JSON envelope for total counts:

```bash
bugsnag errors list --project-id PROJECT_ID --status open | jq '.total_count'
```

### Step 3: Error Trends

```bash
# Daily trend — 2 weeks
bugsnag trends project --project-id PROJECT_ID --resolution 1d --buckets-count 14

# Hourly trend — last 24h
bugsnag trends project --project-id PROJECT_ID --resolution 1h --buckets-count 24
```

`--resolution` and `--buckets-count` are only available on `trends project` (not `trends error`).

Look for:
- **Spikes** — sudden increase, check recent releases
- **Sustained increases** — growing problem
- **Drops after deployments** — fixes working

### Step 4: Stability Metrics

```bash
# Overall stability
bugsnag stability trend --project-id PROJECT_ID

# Production only
bugsnag stability trend --project-id PROJECT_ID --release-stage production
```

`--release-stage` is optional and accepts any stage name (production, staging, development, etc.).

### Step 5: Recent Releases

```bash
bugsnag releases list --project-id PROJECT_ID
```

Compare release timestamps with error trend spikes to identify regressions.

### Step 6: Team Activity

```bash
bugsnag organizations list
bugsnag collaborators list --org-id ORG_ID
```

## Quick Health Check Script

```bash
PROJECT_ID="your-project-id"

bugsnag projects get --project-id "$PROJECT_ID"
bugsnag errors list --project-id "$PROJECT_ID" --status open | jq '.total_count'
bugsnag errors list --project-id "$PROJECT_ID" --status open --severity error --format table
bugsnag trends project --project-id "$PROJECT_ID" --resolution 1d --buckets-count 14
bugsnag stability trend --project-id "$PROJECT_ID" --release-stage production
bugsnag releases list --project-id "$PROJECT_ID" --format table
```

## Interpreting Results

- **Error count trending up** — new bugs introduced, check recent releases
- **Stability dropping** — critical, likely a regression in a recent deploy
- **High event count on a single error** — one error dominating, prioritize it
- **Many open errors with low event count** — noise, consider snoozing old low-impact errors
