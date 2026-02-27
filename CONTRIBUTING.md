# Contributing to bugsnag-cli

Thank you for your interest in contributing! This guide will help you get started.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/<your-username>/bugsnag-cli.git
   cd bugsnag-cli
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   go build -o bugsnag .
   ```

## Development Workflow

1. Create a branch from `main`:
   ```bash
   git checkout -b feature/my-feature
   ```
2. Make your changes
3. Ensure the code compiles and passes checks:
   ```bash
   go build ./...
   go vet ./...
   ```
4. Commit your changes with a clear message
5. Push and open a Pull Request

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `RunE` (not `Run`) on all Cobra commands for proper error propagation
- Return errors instead of calling `os.Exit` directly in commands
- Keep JSON as the default output format — always support both `json` and `table`

## Adding a New Command

1. Create the model in `internal/models/`
   - Implement `TableHeaders()` and `TableRow()` for table output
2. Add the API client method in `internal/client/`
   - Use `FetchSinglePage` / `CollectAllPages` for list endpoints
3. Create the command in `cmd/`
   - Use `RunE`, support `--format`, `--all-pages` where applicable
4. Register the command in its `init()` function

## Reporting Issues

- Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md) template for bugs
- Use the [Feature Request](.github/ISSUE_TEMPLATE/feature_request.md) template for new ideas

## Pull Request Guidelines

- Keep PRs focused — one feature or fix per PR
- Update the README if you add or change commands
- Ensure `go build` and `go vet` pass before submitting

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
