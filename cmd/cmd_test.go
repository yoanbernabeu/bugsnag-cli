package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// resetRootCmd resets all persistent flag values and viper state so that
// each test starts clean.  Because rootCmd is a package-level variable shared
// across tests, we must undo any side-effects from previous executions.
func resetRootCmd() {
	// Reset viper FIRST to avoid leaking state between tests.
	viper.Reset()

	// Reset all local flags on subcommands so values don't leak between tests.
	// This must happen BEFORE we set explicit values below.
	resetSubcommandFlags(rootCmd)

	// Point the config flag at a non-existent file so initConfig does
	// not read the real ~/.bugsnag-cli.yaml.
	rootCmd.PersistentFlags().Set("config", "/dev/null/nonexistent.yaml")

	// Reset all persistent flags to defaults.
	rootCmd.PersistentFlags().Set("api-token", "")
	rootCmd.PersistentFlags().Set("format", "json")
	rootCmd.PersistentFlags().Set("per-page", "30")
	rootCmd.PersistentFlags().Set("all-pages", "false")
	rootCmd.PersistentFlags().Set("base-url", "https://api.bugsnag.com")

	// Also clear any environment variables that might interfere.
	os.Unsetenv("BUGSNAG_API_TOKEN")
	os.Unsetenv("BUGSNAG_FORMAT")
	os.Unsetenv("BUGSNAG_BASE_URL")
	os.Unsetenv("BUGSNAG_PER_PAGE")

	// Re-bind flags to viper since we reset viper.
	viper.BindPFlag("api_token", rootCmd.PersistentFlags().Lookup("api-token"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("per_page", rootCmd.PersistentFlags().Lookup("per-page"))
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))

	viper.SetDefault("format", "json")
	viper.SetDefault("per_page", 30)
	viper.SetDefault("base_url", "https://api.bugsnag.com")
}

// resetSubcommandFlags visits every command in the tree and resets all
// local flags to their default values so that values from a prior test
// do not leak into the next one.
func resetSubcommandFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, child := range cmd.Commands() {
		resetSubcommandFlags(child)
	}
}

// executeCommand runs rootCmd with the given args and captures stdout.
// Returns the captured output and any error from command execution.
func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	_, err := rootCmd.ExecuteC()

	// Cobra writes to the command's OutOrStdout when using Print* helpers,
	// but our code writes to os.Stdout / Printer.Out.  We capture output by
	// temporarily redirecting os.Stdout.
	return buf.String(), err
}

// executeCommandCapture redirects os.Stdout to capture output from Printer,
// which writes to os.Stdout rather than the cobra output buffer.
func executeCommandCapture(args ...string) (string, error) {
	resetRootCmd()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs(args)
	_, err := rootCmd.ExecuteC()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

// newMockServer creates an httptest server that routes requests based on
// the provided handler map. Keys are "METHOD /path" strings.
func newMockServer(handlers map[string]http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		if h, ok := handlers[key]; ok {
			h(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"errors":[{"message":"not found: %s"}]}`, key)
	}))
}

// respondJSON is a helper to write a JSON body.
func respondJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(body)
}

// ---------------------------------------------------------------------------
// Existing tests (preserved)
// ---------------------------------------------------------------------------

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: output.ExitOK,
		},
		{
			name:     "API error type",
			err:      &client.APIError{StatusCode: 401, Message: "Unauthorized"},
			expected: output.ExitAPI,
		},
		{
			name:     "API error string",
			err:      fmt.Errorf("API error (404): Not Found"),
			expected: output.ExitAPI,
		},
		{
			name:     "config error - token",
			err:      fmt.Errorf("API token is required"),
			expected: output.ExitConfig,
		},
		{
			name:     "config error - required flag",
			err:      fmt.Errorf("--project-id is required"),
			expected: output.ExitConfig,
		},
		{
			name:     "network error",
			err:      fmt.Errorf("network error: connection refused"),
			expected: output.ExitNetwork,
		},
		{
			name:     "general error",
			err:      fmt.Errorf("something went wrong"),
			expected: output.ExitGeneral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyError(tt.err)
			if got != tt.expected {
				t.Errorf("classifyError(%v) = %d, want %d", tt.err, got, tt.expected)
			}
		})
	}
}

func TestGetFormatDefault(t *testing.T) {
	// Reset viper for this test
	f := getFormat()
	if f != "json" && f != "" {
		// Default should be json when nothing is set
		t.Logf("format: %s (may vary based on viper state)", f)
	}
}

func TestGetPerPageBounds(t *testing.T) {
	pp := getPerPage()
	if pp <= 0 || pp > 100 {
		t.Errorf("per page out of bounds: %d", pp)
	}
}

func TestGetBaseURLDefault(t *testing.T) {
	url := getBaseURL()
	if url == "" {
		t.Error("base URL should not be empty")
	}
}

// ---------------------------------------------------------------------------
// getAPIToken tests
// ---------------------------------------------------------------------------

func TestGetAPIToken_Set(t *testing.T) {
	resetRootCmd()
	viper.Set("api_token", "test-token-abc")
	defer viper.Reset()

	tok, err := getAPIToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "test-token-abc" {
		t.Errorf("expected 'test-token-abc', got %q", tok)
	}
}

func TestGetAPIToken_NotSet(t *testing.T) {
	resetRootCmd()

	_, err := getAPIToken()
	if err == nil {
		t.Fatal("expected error when token not set")
	}
	if !strings.Contains(err.Error(), "API token is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// getAllPages tests
// ---------------------------------------------------------------------------

func TestGetAllPages_Default(t *testing.T) {
	resetRootCmd()
	if getAllPages() {
		t.Error("expected false by default")
	}
}

func TestGetAllPages_Set(t *testing.T) {
	resetRootCmd()
	rootCmd.PersistentFlags().Set("all-pages", "true")
	if !getAllPages() {
		t.Error("expected true when flag is set")
	}
}

// ---------------------------------------------------------------------------
// Version command
// ---------------------------------------------------------------------------

func TestVersionCommand(t *testing.T) {
	resetRootCmd()
	out, err := executeCommandCapture("version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bugsnag-cli version") {
		t.Errorf("expected version string in output, got: %q", out)
	}
	if !strings.Contains(out, Version) {
		t.Errorf("expected version %q in output, got: %q", Version, out)
	}
}

// ---------------------------------------------------------------------------
// Organizations list
// ---------------------------------------------------------------------------

func TestOrganizationsListCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /user/organizations": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]string{
				{"id": "org1", "name": "My Org", "slug": "my-org", "created_at": "2024-01-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("organizations", "list",
		"--api-token", "tok",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "org1") {
		t.Errorf("expected org1 in output, got: %q", out)
	}
	if !strings.Contains(out, "My Org") {
		t.Errorf("expected 'My Org' in output, got: %q", out)
	}
}

func TestOrganizationsListCommand_NoToken(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("organizations", "list")
	if err == nil {
		t.Fatal("expected error when no token set")
	}
	if !strings.Contains(err.Error(), "API token is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Projects list
// ---------------------------------------------------------------------------

func TestProjectsListCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /organizations/org1/projects": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "p1", "name": "MyProject", "language": "go", "open_error_count": 3, "created_at": "2024-02-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("projects", "list",
		"--api-token", "tok",
		"--org-id", "org1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "p1") {
		t.Errorf("expected p1 in output, got: %q", out)
	}
	if !strings.Contains(out, "MyProject") {
		t.Errorf("expected MyProject in output, got: %q", out)
	}
}

func TestProjectsListCommand_MissingOrgID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("projects", "list",
		"--api-token", "tok")
	if err == nil {
		t.Fatal("expected error when --org-id missing")
	}
	if !strings.Contains(err.Error(), "--org-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectsListCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /organizations/org1/projects": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "p1", "name": "Proj", "language": "go", "open_error_count": 2, "created_at": "2024-01-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("projects", "list",
		"--api-token", "tok",
		"--org-id", "org1",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") {
		t.Errorf("expected table headers in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Projects get
// ---------------------------------------------------------------------------

func TestProjectsGetCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, map[string]any{
				"id": "p1", "name": "Proj", "language": "go",
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("projects", "get",
		"--api-token", "tok",
		"--project-id", "p1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "p1") {
		t.Errorf("expected p1 in output, got: %q", out)
	}
}

func TestProjectsGetCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("projects", "get",
		"--api-token", "tok")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
	if !strings.Contains(err.Error(), "--project-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Errors list
// ---------------------------------------------------------------------------

func TestErrorsListCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/errors": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "e1", "error_class": "TypeError", "severity": "error", "status": "open", "events": 5, "last_seen": "2024-01-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("errors", "list",
		"--api-token", "tok",
		"--project-id", "p1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "e1") {
		t.Errorf("expected e1 in output, got: %q", out)
	}
	if !strings.Contains(out, "TypeError") {
		t.Errorf("expected TypeError in output, got: %q", out)
	}
}

func TestErrorsListCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("errors", "list",
		"--api-token", "tok")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
	if !strings.Contains(err.Error(), "--project-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestErrorsListCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/errors": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "e1", "error_class": "TypeError", "severity": "error", "status": "open", "events": 5, "last_seen": "2024-01-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("errors", "list",
		"--api-token", "tok",
		"--project-id", "p1",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "ERROR_CLASS") {
		t.Errorf("expected table headers in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Errors get
// ---------------------------------------------------------------------------

func TestErrorsGetCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/errors/e1": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, map[string]any{
				"id": "e1", "error_class": "SyntaxError", "severity": "warning",
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("errors", "get",
		"--api-token", "tok",
		"--project-id", "p1",
		"--error-id", "e1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "SyntaxError") {
		t.Errorf("expected SyntaxError in output, got: %q", out)
	}
}

func TestErrorsGetCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("errors", "get",
		"--api-token", "tok",
		"--error-id", "e1")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
	if !strings.Contains(err.Error(), "--project-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestErrorsGetCommand_MissingErrorID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("errors", "get",
		"--api-token", "tok",
		"--project-id", "p1")
	if err == nil {
		t.Fatal("expected error when --error-id missing")
	}
	if !strings.Contains(err.Error(), "--error-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Events list
// ---------------------------------------------------------------------------

func TestEventsListCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/events": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "ev1", "error_class": "Error", "severity": "info", "context": "/api", "received_at": "2024-01-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("events", "list",
		"--api-token", "tok",
		"--project-id", "p1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ev1") {
		t.Errorf("expected ev1 in output, got: %q", out)
	}
}

func TestEventsListCommand_WithErrorID(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/errors/e1/events": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "ev2", "error_class": "Error", "severity": "warning"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("events", "list",
		"--api-token", "tok",
		"--project-id", "p1",
		"--error-id", "e1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ev2") {
		t.Errorf("expected ev2 in output, got: %q", out)
	}
}

func TestEventsListCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("events", "list",
		"--api-token", "tok")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
	if !strings.Contains(err.Error(), "--project-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Events get
// ---------------------------------------------------------------------------

func TestEventsGetCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/events/ev1": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, map[string]any{
				"id": "ev1", "error_class": "NullPointer", "severity": "error",
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("events", "get",
		"--api-token", "tok",
		"--project-id", "p1",
		"--event-id", "ev1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "NullPointer") {
		t.Errorf("expected NullPointer in output, got: %q", out)
	}
}

func TestEventsGetCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("events", "get",
		"--api-token", "tok",
		"--event-id", "ev1")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
}

func TestEventsGetCommand_MissingEventID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("events", "get",
		"--api-token", "tok",
		"--project-id", "p1")
	if err == nil {
		t.Fatal("expected error when --event-id missing")
	}
	if !strings.Contains(err.Error(), "--event-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Trends project
// ---------------------------------------------------------------------------

func TestTrendsProjectCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/trend": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"from": "2024-01-01", "to": "2024-01-02", "events_count": 42},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("trends", "project",
		"--api-token", "tok",
		"--project-id", "p1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("expected 42 in output, got: %q", out)
	}
}

func TestTrendsProjectCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("trends", "project",
		"--api-token", "tok")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
	if !strings.Contains(err.Error(), "--project-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTrendsProjectCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/trend": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"from": "2024-01-01", "to": "2024-01-02", "events_count": 10},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("trends", "project",
		"--api-token", "tok",
		"--project-id", "p1",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "FROM") || !strings.Contains(out, "TO") {
		t.Errorf("expected table headers in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Trends error
// ---------------------------------------------------------------------------

func TestTrendsErrorCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/errors/e1/trend": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"from": "2024-01-01", "to": "2024-01-02", "events_count": 7},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("trends", "error",
		"--api-token", "tok",
		"--project-id", "p1",
		"--error-id", "e1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "7") {
		t.Errorf("expected 7 in output, got: %q", out)
	}
}

func TestTrendsErrorCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("trends", "error",
		"--api-token", "tok",
		"--error-id", "e1")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
}

func TestTrendsErrorCommand_MissingErrorID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("trends", "error",
		"--api-token", "tok",
		"--project-id", "p1")
	if err == nil {
		t.Fatal("expected error when --error-id missing")
	}
	if !strings.Contains(err.Error(), "--error-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Collaborators list
// ---------------------------------------------------------------------------

func TestCollaboratorsListCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /organizations/org1/collaborators": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "c1", "name": "Alice", "email": "alice@test.com", "is_admin": true},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("collaborators", "list",
		"--api-token", "tok",
		"--org-id", "org1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected Alice in output, got: %q", out)
	}
}

func TestCollaboratorsListCommand_MissingOrgID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("collaborators", "list",
		"--api-token", "tok")
	if err == nil {
		t.Fatal("expected error when --org-id missing")
	}
	if !strings.Contains(err.Error(), "--org-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCollaboratorsListCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /organizations/org1/collaborators": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "c1", "name": "Alice", "email": "alice@test.com", "is_admin": true},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("collaborators", "list",
		"--api-token", "tok",
		"--org-id", "org1",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "IS_ADMIN") {
		t.Errorf("expected IS_ADMIN header in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Comments list
// ---------------------------------------------------------------------------

func TestCommentsListCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/errors/e1/comments": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "cm1", "message": "looks bad", "author_name": "Bob", "created_at": "2024-01-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("comments", "list",
		"--api-token", "tok",
		"--project-id", "p1",
		"--error-id", "e1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "looks bad") {
		t.Errorf("expected comment message in output, got: %q", out)
	}
}

func TestCommentsListCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("comments", "list",
		"--api-token", "tok",
		"--error-id", "e1")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
}

func TestCommentsListCommand_MissingErrorID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("comments", "list",
		"--api-token", "tok",
		"--project-id", "p1")
	if err == nil {
		t.Fatal("expected error when --error-id missing")
	}
	if !strings.Contains(err.Error(), "--error-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Comments create
// ---------------------------------------------------------------------------

func TestCommentsCreateCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"POST /projects/p1/errors/e1/comments": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 201, map[string]any{
				"id": "cm2", "message": "fixed it", "author_name": "Alice", "created_at": "2024-02-01",
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("comments", "create",
		"--api-token", "tok",
		"--project-id", "p1",
		"--error-id", "e1",
		"--message", "fixed it",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "fixed it") {
		t.Errorf("expected 'fixed it' in output, got: %q", out)
	}
}

func TestCommentsCreateCommand_MissingMessage(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("comments", "create",
		"--api-token", "tok",
		"--project-id", "p1",
		"--error-id", "e1")
	if err == nil {
		t.Fatal("expected error when --message missing")
	}
	if !strings.Contains(err.Error(), "--message is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommentsCreateCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("comments", "create",
		"--api-token", "tok",
		"--error-id", "e1",
		"--message", "hello")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
}

func TestCommentsCreateCommand_MissingErrorID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("comments", "create",
		"--api-token", "tok",
		"--project-id", "p1",
		"--message", "hello")
	if err == nil {
		t.Fatal("expected error when --error-id missing")
	}
}

// ---------------------------------------------------------------------------
// Releases list
// ---------------------------------------------------------------------------

func TestReleasesListCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/releases": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "r1", "app_version": "1.0.0", "release_stage": map[string]string{"name": "production"}, "release_source": "api", "release_time": "2024-01-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("releases", "list",
		"--api-token", "tok",
		"--project-id", "p1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "1.0.0") {
		t.Errorf("expected 1.0.0 in output, got: %q", out)
	}
}

func TestReleasesListCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("releases", "list",
		"--api-token", "tok")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
	if !strings.Contains(err.Error(), "--project-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReleasesListCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/releases": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "r1", "app_version": "2.0", "release_stage": map[string]string{"name": "staging"}, "release_source": "api", "release_time": "2024-06-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("releases", "list",
		"--api-token", "tok",
		"--project-id", "p1",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "VERSION") || !strings.Contains(out, "RELEASE_STAGE") {
		t.Errorf("expected table headers in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Stability trend
// ---------------------------------------------------------------------------

func TestStabilityTrendCommand(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/stability_trend": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, map[string]any{
				"project_id":       "p1",
				"release_stage_name": "production",
				"timeline_points": []map[string]any{
					{"bucket_start": "2024-01-01", "bucket_end": "2024-01-02", "total_sessions_count": 100, "unhandled_sessions_count": 5, "unhandled_rate": 0.05},
				},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("stability", "trend",
		"--api-token", "tok",
		"--project-id", "p1",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "p1") {
		t.Errorf("expected p1 in output, got: %q", out)
	}
}

func TestStabilityTrendCommand_MissingProjectID(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture("stability", "trend",
		"--api-token", "tok")
	if err == nil {
		t.Fatal("expected error when --project-id missing")
	}
	if !strings.Contains(err.Error(), "--project-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStabilityTrendCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/stability_trend": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, map[string]any{
				"project_id":       "p1",
				"release_stage_name": "production",
				"timeline_points": []map[string]any{
					{"bucket_start": "2024-01-01", "bucket_end": "2024-01-02", "total_sessions_count": 100, "unhandled_sessions_count": 5, "unhandled_rate": 0.05},
				},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("stability", "trend",
		"--api-token", "tok",
		"--project-id", "p1",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "BUCKET_START") || !strings.Contains(out, "SESSIONS") {
		t.Errorf("expected table headers in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// --all-pages flag (pagination)
// ---------------------------------------------------------------------------

func TestOrganizationsListCommand_AllPages(t *testing.T) {
	resetRootCmd()
	callCount := 0
	// We need the server URL in the Link header, but we do not know it
	// until the server is created.  Use a variable to capture it.
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First page: include Link header pointing to next page on the same server.
			w.Header().Set("Link", fmt.Sprintf(`<%s/user/organizations?page=2&per_page=30>; rel="next"`, srvURL))
			respondJSON(w, 200, []map[string]string{
				{"id": "org1", "name": "First"},
			})
		} else {
			// Second (and last) page -- no Link header.
			respondJSON(w, 200, []map[string]string{
				{"id": "org2", "name": "Second"},
			})
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	out, err := executeCommandCapture("organizations", "list",
		"--api-token", "tok",
		"--all-pages",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain data from both pages
	if !strings.Contains(out, "org1") {
		t.Errorf("expected org1 in output, got: %q", out)
	}
	if !strings.Contains(out, "org2") {
		t.Errorf("expected org2 in output, got: %q", out)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (pagination), got %d", callCount)
	}
}

// ---------------------------------------------------------------------------
// Configure command
// ---------------------------------------------------------------------------

func TestConfigureCommand(t *testing.T) {
	resetRootCmd()
	// configure writes to ~/.bugsnag-cli.yaml which we do not want to
	// modify in tests.  We cannot easily redirect it without modifying
	// the production code, so we just test the missing-token case.
	_, err := executeCommandCapture("configure")
	if err == nil {
		t.Fatal("expected error when --api-token is missing")
	}
	if !strings.Contains(err.Error(), "--api-token is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConfigureCommand_WritesConfig(t *testing.T) {
	// This test creates a temp directory and overrides HOME so that
	// configure writes to a temp file instead of the real home dir.
	resetRootCmd()

	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	out, err := executeCommandCapture("configure",
		"--api-token", "my-secret-token",
		"--default-format", "table",
		"--default-per-page", "50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfgPath := filepath.Join(tmpDir, ".bugsnag-cli.yaml")
	data, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		t.Fatalf("config file not created: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "api_token: my-secret-token") {
		t.Errorf("expected api_token in config, got: %q", content)
	}
	if !strings.Contains(content, "format: table") {
		t.Errorf("expected format in config, got: %q", content)
	}
	if !strings.Contains(content, "per_page: 50") {
		t.Errorf("expected per_page in config, got: %q", content)
	}
	// Output should contain JSON with status=ok
	if !strings.Contains(out, "ok") {
		t.Errorf("expected 'ok' in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// API error handling (server returns error)
// ---------------------------------------------------------------------------

func TestCommand_APIError(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /user/organizations": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 401, map[string]any{
				"errors": []map[string]string{{"message": "Unauthorized"}},
			})
		},
	})
	defer srv.Close()

	_, err := executeCommandCapture("organizations", "list",
		"--api-token", "bad-tok",
		"--base-url", srv.URL)
	if err == nil {
		t.Fatal("expected error on 401")
	}
}

// ---------------------------------------------------------------------------
// getPerPage edge cases
// ---------------------------------------------------------------------------

func TestGetPerPage_CustomValue(t *testing.T) {
	resetRootCmd()
	viper.Set("per_page", 50)
	defer viper.Reset()

	pp := getPerPage()
	if pp != 50 {
		t.Errorf("expected 50, got %d", pp)
	}
}

func TestGetPerPage_OutOfBoundsHigh(t *testing.T) {
	resetRootCmd()
	viper.Set("per_page", 200)
	defer viper.Reset()

	pp := getPerPage()
	if pp != 30 {
		t.Errorf("expected default 30 for out-of-bounds, got %d", pp)
	}
}

func TestGetPerPage_OutOfBoundsZero(t *testing.T) {
	resetRootCmd()
	viper.Set("per_page", 0)
	defer viper.Reset()

	pp := getPerPage()
	if pp != 30 {
		t.Errorf("expected default 30 for zero, got %d", pp)
	}
}

// ---------------------------------------------------------------------------
// getBaseURL custom value
// ---------------------------------------------------------------------------

func TestGetBaseURL_Custom(t *testing.T) {
	resetRootCmd()
	viper.Set("base_url", "https://custom.api.com")
	defer viper.Reset()

	u := getBaseURL()
	if u != "https://custom.api.com" {
		t.Errorf("expected custom URL, got %q", u)
	}
}

// ---------------------------------------------------------------------------
// getFormat custom value
// ---------------------------------------------------------------------------

func TestGetFormat_Table(t *testing.T) {
	resetRootCmd()
	viper.Set("format", "table")
	defer viper.Reset()

	f := getFormat()
	if f != "table" {
		t.Errorf("expected table, got %q", f)
	}
}

func TestGetFormat_EmptyDefaultsToJSON(t *testing.T) {
	resetRootCmd()
	viper.Set("format", "")
	defer viper.Reset()

	f := getFormat()
	if f != "json" {
		t.Errorf("expected json for empty format, got %q", f)
	}
}

// ---------------------------------------------------------------------------
// classifyError with cobra.Command (ensure subcommand errors go through)
// ---------------------------------------------------------------------------

func TestClassifyError_WrappedAPIError(t *testing.T) {
	apiErr := &client.APIError{StatusCode: 500, Message: "Internal Server Error"}
	code := classifyError(apiErr)
	if code != output.ExitAPI {
		t.Errorf("expected ExitAPI, got %d", code)
	}
}

// ---------------------------------------------------------------------------
// Root command usage (no subcommand)
// ---------------------------------------------------------------------------

func TestRootCommand_NoArgs(t *testing.T) {
	resetRootCmd()
	_, err := executeCommandCapture()
	// Running without subcommand should not error -- cobra just prints help.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Verify subcommand registration
// ---------------------------------------------------------------------------

func TestSubcommandRegistration(t *testing.T) {
	names := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		names[c.Name()] = true
	}

	expected := []string{
		"version", "configure", "organizations", "projects",
		"errors", "events", "trends", "collaborators",
		"comments", "releases", "stability",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Ensure commands have correct Use field (useful for CLI parsing)
// ---------------------------------------------------------------------------

func TestRootCmdUseName(t *testing.T) {
	if rootCmd.Use != "bugsnag" {
		t.Errorf("expected root command Use='bugsnag', got %q", rootCmd.Use)
	}
}

// ---------------------------------------------------------------------------
// Test that --format flag is wired correctly for organizations list
// ---------------------------------------------------------------------------

func TestOrganizationsListCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /user/organizations": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]string{
				{"id": "org1", "name": "My Org", "slug": "my-org", "created_at": "2024-01-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("organizations", "list",
		"--api-token", "tok",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "SLUG") {
		t.Errorf("expected table headers in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Test events list table format
// ---------------------------------------------------------------------------

func TestEventsListCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p2/events": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "ev1", "error_class": "RangeError", "severity": "info", "context": "/home", "received_at": "2024-05-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("events", "list",
		"--api-token", "tok",
		"--project-id", "p2",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ERROR_CLASS") || !strings.Contains(out, "SEVERITY") {
		t.Errorf("expected table headers in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Test comments list table format
// ---------------------------------------------------------------------------

func TestCommentsListCommand_TableFormat(t *testing.T) {
	resetRootCmd()
	srv := newMockServer(map[string]http.HandlerFunc{
		"GET /projects/p1/errors/e1/comments": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, []map[string]any{
				{"id": "cm1", "message": "need to fix", "author_name": "Carol", "created_at": "2024-03-01"},
			})
		},
	})
	defer srv.Close()

	out, err := executeCommandCapture("comments", "list",
		"--api-token", "tok",
		"--project-id", "p1",
		"--error-id", "e1",
		"--format", "table",
		"--base-url", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "AUTHOR") || !strings.Contains(out, "MESSAGE") {
		t.Errorf("expected table headers in output, got: %q", out)
	}
}

// ---------------------------------------------------------------------------
// Ensure cobra.Command returned by rootCmd.Commands() has expected children
// ---------------------------------------------------------------------------

func TestTrendsSubcommands(t *testing.T) {
	var trendsC *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "trends" {
			trendsC = c
			break
		}
	}
	if trendsC == nil {
		t.Fatal("trends command not found")
	}

	names := make(map[string]bool)
	for _, c := range trendsC.Commands() {
		names[c.Name()] = true
	}
	if !names["project"] {
		t.Error("expected 'project' subcommand under trends")
	}
	if !names["error"] {
		t.Error("expected 'error' subcommand under trends")
	}
}

func TestCommentsSubcommands(t *testing.T) {
	var commentsC *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "comments" {
			commentsC = c
			break
		}
	}
	if commentsC == nil {
		t.Fatal("comments command not found")
	}

	names := make(map[string]bool)
	for _, c := range commentsC.Commands() {
		names[c.Name()] = true
	}
	if !names["list"] {
		t.Error("expected 'list' subcommand under comments")
	}
	if !names["create"] {
		t.Error("expected 'create' subcommand under comments")
	}
}
