# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `organizations list` command
- `projects list|get` commands
- `errors list|get` commands with filtering (status, severity, sort, direction)
- `events list|get` commands with optional error scoping
- `trends project|error` commands
- `collaborators list` command
- `comments list|create` commands
- `releases list` command
- `stability trend` command
- `configure` command to save API token to `~/.bugsnag-cli.yaml`
- `version` command
- JSON output by default with `{"data": [...], "total_count": N, "has_more": bool}` envelope
- Table output via `--format table`
- Auto-pagination with `--all-pages`
- Configuration via flags, environment variables (`BUGSNAG_*`), or config file
- Structured error output on stderr with exit codes (0-4)
- GoReleaser CI for multi-platform releases
