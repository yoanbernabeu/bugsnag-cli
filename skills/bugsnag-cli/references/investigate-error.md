# Investigate a Bugsnag Error

Complete error investigation workflow — from "something is broken" to understanding the error's scope, history, and details.

## Prerequisites

- bugsnag-cli configured with a valid API token
- A project ID (run `bugsnag projects list --org-id ORG_ID` to find it)

## Investigation Workflow

### Step 1: Find the Error

List errors with filters to narrow down what you're looking for:

```bash
# All open errors, sorted by most recent
bugsnag errors list --project-id PROJECT_ID --status open --sort last_seen --direction desc

# Only critical errors
bugsnag errors list --project-id PROJECT_ID --severity error --status open

# All errors across every page
bugsnag errors list --project-id PROJECT_ID --all-pages
```

Available filters:
- `--status`: open, fixed, snoozed, ignored
- `--severity`: info, warning, error
- `--sort`: created_at, last_seen, events, users, unsorted
- `--direction`: asc, desc

### Step 2: Get Error Details

```bash
bugsnag errors get --project-id PROJECT_ID --error-id ERROR_ID
```

Returns full error details: class, message, context, severity, status, metadata.

### Step 3: Examine Events (Occurrences)

Each error groups multiple event occurrences. Events contain stack traces, request data, device info, and custom metadata — this is where the actual debugging details live.

```bash
# List events for this specific error
bugsnag events list --project-id PROJECT_ID --error-id ERROR_ID

# Get full details of a specific event
bugsnag events get --project-id PROJECT_ID --event-id EVENT_ID
```

### Step 4: Check Error Trends

```bash
bugsnag trends error --project-id PROJECT_ID --error-id ERROR_ID
```

Shows event counts bucketed by time period. Look for:
- **Spikes** — sudden increase, possibly tied to a deploy
- **Sustained increases** — growing problem
- **Drops** — a fix is working

**Important:** `trends error` does NOT accept `--resolution` or `--buckets-count` (only `trends project` does).

### Step 5: Read Existing Comments

```bash
bugsnag comments list --project-id PROJECT_ID --error-id ERROR_ID
```

### Step 6: Document Your Findings

```bash
bugsnag comments create --project-id PROJECT_ID --error-id ERROR_ID \
  --message "Investigated: root cause is X. Fix planned in PR #123."
```

## Quick Investigation Script

Gather all info at once:

```bash
PROJECT_ID="your-project-id"
ERROR_ID="your-error-id"

# Error overview
bugsnag errors get --project-id "$PROJECT_ID" --error-id "$ERROR_ID"

# Recent events (stack traces)
bugsnag events list --project-id "$PROJECT_ID" --error-id "$ERROR_ID"

# Trend
bugsnag trends error --project-id "$PROJECT_ID" --error-id "$ERROR_ID"

# Existing comments
bugsnag comments list --project-id "$PROJECT_ID" --error-id "$ERROR_ID"
```

## Tips

- Use `--format table` for quick human review during investigation
- Use JSON (default) when parsing results programmatically
- Event details are the most valuable for debugging — they contain stack traces
- `--all-pages` is useful when an error has many events to review
