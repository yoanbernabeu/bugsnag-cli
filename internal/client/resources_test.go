package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

// ---------------------------------------------------------------------------
// Helper: newTestServer creates an httptest.Server that serves the given JSON
// payload and optionally sets a Link header for pagination.
// ---------------------------------------------------------------------------

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func jsonHandler(t *testing.T, statusCode int, payload any) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if payload != nil {
			if err := json.NewEncoder(w).Encode(payload); err != nil {
				t.Fatalf("encoding response: %v", err)
			}
		}
	}
}

func errorHandler(statusCode int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{{"message": message}},
		})
	}
}

// ---------------------------------------------------------------------------
// newRequestWithBody
// ---------------------------------------------------------------------------

func TestNewRequestWithBody(t *testing.T) {
	c := New("https://api.bugsnag.com", "test-token", 30)
	bodyStr := `{"message":"hello"}`
	req, err := c.newRequestWithBody("POST", "/some/path", strings.NewReader(bodyStr))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Method != "POST" {
		t.Errorf("expected POST, got %s", req.Method)
	}
	if req.URL.String() != "https://api.bugsnag.com/some/path" {
		t.Errorf("unexpected URL: %s", req.URL.String())
	}
	if req.Header.Get("Authorization") != "token test-token" {
		t.Errorf("unexpected Authorization header: %s", req.Header.Get("Authorization"))
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("unexpected Content-Type: %s", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("unexpected Accept header: %s", req.Header.Get("Accept"))
	}

	read, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	if string(read) != bodyStr {
		t.Errorf("expected body %q, got %q", bodyStr, string(read))
	}
}

// ===========================================================================
// Organizations
// ===========================================================================

func TestListOrganizations_SinglePage(t *testing.T) {
	orgs := []models.Organization{
		{ID: "org-1", Name: "Org One", Slug: "org-one", CreatedAt: "2024-01-01"},
		{ID: "org-2", Name: "Org Two", Slug: "org-two", CreatedAt: "2024-02-01"},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/organizations" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orgs)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListOrganizations(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 orgs, got %d", len(result))
	}
	if result[0].ID != "org-1" || result[0].Name != "Org One" {
		t.Errorf("unexpected org[0]: %+v", result[0])
	}
	if result[1].Slug != "org-two" {
		t.Errorf("unexpected org[1].Slug: %s", result[1].Slug)
	}
}

func TestListOrganizations_SinglePageWithPagination(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", `<http://`+r.Host+`/user/organizations?offset=next>; rel="next"`)
		json.NewEncoder(w).Encode([]models.Organization{
			{ID: "org-1", Name: "Org One"},
		})
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListOrganizations(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true when Link header is present")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 org, got %d", len(result))
	}
}

func TestListOrganizations_AllPages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/user/organizations?offset=page2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Organization{
				{ID: "org-1", Name: "Org One"},
			})
		} else {
			json.NewEncoder(w).Encode([]models.Organization{
				{ID: "org-2", Name: "Org Two"},
			})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListOrganizations(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("allPages mode should return hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 orgs across pages, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestListOrganizations_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(401, "Unauthorized"))
	defer server.Close()

	c := New(server.URL, "bad-token", 30)
	_, _, err := c.ListOrganizations(false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Projects
// ===========================================================================

func TestListProjects_SinglePage(t *testing.T) {
	projects := []models.Project{
		{ID: "proj-1", Name: "My Project", Language: "go", OpenErrorCount: 5, CreatedAt: "2024-01-01"},
		{ID: "proj-2", Name: "Other Project", Language: "python", OpenErrorCount: 12, CreatedAt: "2024-03-01"},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/organizations/org-abc/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListProjects("org-abc", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(result))
	}
	if result[0].Name != "My Project" || result[0].Language != "go" {
		t.Errorf("unexpected project[0]: %+v", result[0])
	}
	if result[1].OpenErrorCount != 12 {
		t.Errorf("expected OpenErrorCount=12, got %d", result[1].OpenErrorCount)
	}
}

func TestListProjects_AllPages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/organizations/org-abc/projects?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Project{{ID: "proj-1", Name: "P1"}})
		} else {
			json.NewEncoder(w).Encode([]models.Project{{ID: "proj-2", Name: "P2"}})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListProjects("org-abc", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("allPages should return hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestListProjects_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Organization not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, _, err := c.ListProjects("bad-org", false)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Organization not found" {
		t.Errorf("unexpected message: %s", apiErr.Message)
	}
}

func TestGetProject_Success(t *testing.T) {
	project := models.Project{
		ID:             "proj-123",
		Name:           "Test Project",
		Slug:           "test-project",
		APIKey:         "api-key-abc",
		Language:       "javascript",
		OpenErrorCount: 42,
		CreatedAt:      "2024-01-15",
		URL:            "https://api.bugsnag.com/projects/proj-123",
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(project)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetProject("proj-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "proj-123" {
		t.Errorf("expected ID=proj-123, got %s", result.ID)
	}
	if result.Name != "Test Project" {
		t.Errorf("expected Name='Test Project', got %s", result.Name)
	}
	if result.Language != "javascript" {
		t.Errorf("expected Language=javascript, got %s", result.Language)
	}
	if result.OpenErrorCount != 42 {
		t.Errorf("expected OpenErrorCount=42, got %d", result.OpenErrorCount)
	}
	if result.APIKey != "api-key-abc" {
		t.Errorf("expected APIKey=api-key-abc, got %s", result.APIKey)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Project not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetProject("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestGetProject_ServerError(t *testing.T) {
	server := newTestServer(t, errorHandler(500, "Internal Server Error"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetProject("proj-123")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Errors
// ===========================================================================

func TestListErrors_SinglePage(t *testing.T) {
	errors := []models.BugsnagError{
		{
			ID:         "err-1",
			ProjectID:  "proj-1",
			ErrorClass: "NullPointerException",
			Message:    "null ref",
			Severity:   "error",
			Status:     "open",
			EventsCount: 100,
			LastSeen:   "2024-06-01",
		},
		{
			ID:         "err-2",
			ProjectID:  "proj-1",
			ErrorClass: "RuntimeError",
			Severity:   "warning",
			Status:     "fixed",
			EventsCount: 5,
			LastSeen:   "2024-05-15",
		},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/errors" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errors)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListErrors(ListErrorsOptions{ProjectID: "proj-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(result))
	}
	if result[0].ErrorClass != "NullPointerException" {
		t.Errorf("unexpected ErrorClass: %s", result[0].ErrorClass)
	}
	if result[1].Status != "fixed" {
		t.Errorf("expected status=fixed, got %s", result[1].Status)
	}
}

func TestListErrors_WithFilters(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("status") != "open" {
			t.Errorf("expected status=open, got %s", q.Get("status"))
		}
		if q.Get("severity") != "error" {
			t.Errorf("expected severity=error, got %s", q.Get("severity"))
		}
		if q.Get("sort") != "last_seen" {
			t.Errorf("expected sort=last_seen, got %s", q.Get("sort"))
		}
		if q.Get("direction") != "desc" {
			t.Errorf("expected direction=desc, got %s", q.Get("direction"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.BugsnagError{
			{ID: "err-1", Status: "open", Severity: "error"},
		})
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, _, err := c.ListErrors(ListErrorsOptions{
		ProjectID: "proj-1",
		Status:    "open",
		Severity:  "error",
		Sort:      "last_seen",
		Direction: "desc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result))
	}
}

func TestListErrors_AllPages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/projects/proj-1/errors?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.BugsnagError{{ID: "err-1"}})
		} else {
			json.NewEncoder(w).Encode([]models.BugsnagError{{ID: "err-2"}})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListErrors(ListErrorsOptions{
		ProjectID: "proj-1",
		AllPages:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("allPages should return hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestListErrors_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Project not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, _, err := c.ListErrors(ListErrorsOptions{ProjectID: "bad-proj"})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestGetError_Success(t *testing.T) {
	bugErr := models.BugsnagError{
		ID:         "err-123",
		ProjectID:  "proj-1",
		ErrorClass: "TypeError",
		Message:    "Cannot read property 'x' of undefined",
		Severity:   "error",
		Status:     "open",
		Unhandled:  true,
		EventsCount: 250,
		FirstSeen:  "2024-01-10",
		LastSeen:   "2024-06-20",
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/errors/err-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bugErr)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetError("proj-1", "err-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "err-123" {
		t.Errorf("expected ID=err-123, got %s", result.ID)
	}
	if result.ErrorClass != "TypeError" {
		t.Errorf("expected ErrorClass=TypeError, got %s", result.ErrorClass)
	}
	if result.Message != "Cannot read property 'x' of undefined" {
		t.Errorf("unexpected Message: %s", result.Message)
	}
	if !result.Unhandled {
		t.Error("expected Unhandled=true")
	}
	if result.EventsCount != 250 {
		t.Errorf("expected EventsCount=250, got %d", result.EventsCount)
	}
}

func TestGetError_NotFound(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Error not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetError("proj-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestGetError_ServerError(t *testing.T) {
	server := newTestServer(t, errorHandler(500, "Internal Server Error"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetError("proj-1", "err-123")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Events
// ===========================================================================

func TestListEvents_ForProject_SinglePage(t *testing.T) {
	events := []models.Event{
		{ID: "evt-1", ProjectID: "proj-1", ErrorClass: "TypeError", Severity: "error", ReceivedAt: "2024-06-01T10:00:00Z"},
		{ID: "evt-2", ProjectID: "proj-1", ErrorClass: "RangeError", Severity: "warning", ReceivedAt: "2024-06-01T11:00:00Z"},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListEvents("proj-1", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}
	if result[0].ID != "evt-1" {
		t.Errorf("expected ID=evt-1, got %s", result[0].ID)
	}
}

func TestListEvents_ForError_SinglePage(t *testing.T) {
	events := []models.Event{
		{ID: "evt-1", ErrorID: "err-1"},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/errors/err-1/events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, _, err := c.ListEvents("proj-1", "err-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result))
	}
	if result[0].ErrorID != "err-1" {
		t.Errorf("expected ErrorID=err-1, got %s", result[0].ErrorID)
	}
}

func TestListEvents_AllPages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/projects/proj-1/events?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Event{{ID: "evt-1"}})
		} else {
			json.NewEncoder(w).Encode([]models.Event{{ID: "evt-2"}, {ID: "evt-3"}})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListEvents("proj-1", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("allPages should return hasMore=false")
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 events, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestListEvents_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(500, "Server exploded"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, _, err := c.ListEvents("proj-1", "", false)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

func TestGetEvent_Success(t *testing.T) {
	event := models.Event{
		ID:         "evt-abc",
		ProjectID:  "proj-1",
		ErrorID:    "err-1",
		ErrorClass: "NullPointerException",
		Message:    "null",
		Severity:   "error",
		Unhandled:  true,
		Context:    "com.example.App",
		ReceivedAt: "2024-06-15T12:00:00Z",
		URL:        "https://app.bugsnag.com/events/evt-abc",
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/events/evt-abc" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(event)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetEvent("proj-1", "evt-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "evt-abc" {
		t.Errorf("expected ID=evt-abc, got %s", result.ID)
	}
	if result.ErrorClass != "NullPointerException" {
		t.Errorf("expected ErrorClass=NullPointerException, got %s", result.ErrorClass)
	}
	if !result.Unhandled {
		t.Error("expected Unhandled=true")
	}
	if result.Context != "com.example.App" {
		t.Errorf("expected Context=com.example.App, got %s", result.Context)
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Event not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetEvent("proj-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestGetEvent_ServerError(t *testing.T) {
	server := newTestServer(t, errorHandler(500, "Internal failure"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetEvent("proj-1", "evt-abc")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Trends
// ===========================================================================

func TestGetProjectTrends_Success(t *testing.T) {
	buckets := []models.TrendBucket{
		{From: "2024-06-01", To: "2024-06-02", EventsCount: 10},
		{From: "2024-06-02", To: "2024-06-03", EventsCount: 25},
		{From: "2024-06-03", To: "2024-06-04", EventsCount: 3},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/trend" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("resolution") != "1d" {
			t.Errorf("expected resolution=1d, got %s", q.Get("resolution"))
		}
		if q.Get("buckets_count") != "3" {
			t.Errorf("expected buckets_count=3, got %s", q.Get("buckets_count"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buckets)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetProjectTrends("proj-1", "1d", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(result))
	}
	if result[0].EventsCount != 10 {
		t.Errorf("expected bucket[0].EventsCount=10, got %d", result[0].EventsCount)
	}
	if result[1].From != "2024-06-02" {
		t.Errorf("expected bucket[1].From=2024-06-02, got %s", result[1].From)
	}
	if result[2].To != "2024-06-04" {
		t.Errorf("expected bucket[2].To=2024-06-04, got %s", result[2].To)
	}
}

func TestGetProjectTrends_DefaultParams(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("resolution") != "" {
			t.Errorf("expected no resolution param, got %s", q.Get("resolution"))
		}
		if q.Get("buckets_count") != "" {
			t.Errorf("expected no buckets_count param, got %s", q.Get("buckets_count"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.TrendBucket{})
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetProjectTrends("proj-1", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 buckets, got %d", len(result))
	}
}

func TestGetProjectTrends_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Project not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetProjectTrends("bad-proj", "1d", 5)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestGetErrorTrends_Success(t *testing.T) {
	buckets := []models.TrendBucket{
		{From: "2024-06-01", To: "2024-06-08", EventsCount: 42},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/errors/err-1/trend" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buckets)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetErrorTrends("proj-1", "err-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(result))
	}
	if result[0].EventsCount != 42 {
		t.Errorf("expected EventsCount=42, got %d", result[0].EventsCount)
	}
}

func TestGetErrorTrends_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(500, "Internal Server Error"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetErrorTrends("proj-1", "err-1")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Collaborators
// ===========================================================================

func TestListCollaborators_SinglePage(t *testing.T) {
	collabs := []models.Collaborator{
		{ID: "user-1", Email: "alice@example.com", Name: "Alice", IsAdmin: true, ProjectsCount: 5, CreatedAt: "2024-01-01"},
		{ID: "user-2", Email: "bob@example.com", Name: "Bob", IsAdmin: false, ProjectsCount: 2, CreatedAt: "2024-02-01"},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/organizations/org-1/collaborators" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collabs)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListCollaborators("org-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 collaborators, got %d", len(result))
	}
	if result[0].Name != "Alice" || !result[0].IsAdmin {
		t.Errorf("unexpected collaborator[0]: %+v", result[0])
	}
	if result[1].Email != "bob@example.com" {
		t.Errorf("expected Email=bob@example.com, got %s", result[1].Email)
	}
}

func TestListCollaborators_AllPages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/organizations/org-1/collaborators?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Collaborator{{ID: "user-1", Name: "Alice"}})
		} else {
			json.NewEncoder(w).Encode([]models.Collaborator{{ID: "user-2", Name: "Bob"}})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListCollaborators("org-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("allPages should return hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 collaborators, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestListCollaborators_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Organization not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, _, err := c.ListCollaborators("bad-org", false)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Comments
// ===========================================================================

func TestListComments_SinglePage(t *testing.T) {
	comments := []models.Comment{
		{ID: "cmt-1", Message: "First comment", AuthorID: "user-1", AuthorName: "Alice", CreatedAt: "2024-06-01T10:00:00Z"},
		{ID: "cmt-2", Message: "Second comment", AuthorID: "user-2", AuthorName: "Bob", CreatedAt: "2024-06-02T11:00:00Z"},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/errors/err-1/comments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListComments("proj-1", "err-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result))
	}
	if result[0].Message != "First comment" {
		t.Errorf("unexpected message: %s", result[0].Message)
	}
	if result[1].AuthorName != "Bob" {
		t.Errorf("expected AuthorName=Bob, got %s", result[1].AuthorName)
	}
}

func TestListComments_AllPages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/projects/proj-1/errors/err-1/comments?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Comment{{ID: "cmt-1", Message: "First"}})
		} else {
			json.NewEncoder(w).Encode([]models.Comment{{ID: "cmt-2", Message: "Second"}})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListComments("proj-1", "err-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("allPages should return hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestListComments_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Error not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, _, err := c.ListComments("proj-1", "bad-err", false)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestCreateComment_Success(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/errors/err-1/comments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content-type: %s", r.Header.Get("Content-Type"))
		}

		// Verify request body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		var body map[string]string
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("unmarshaling request body: %v", err)
		}
		if body["message"] != "This is a test comment" {
			t.Errorf("expected message='This is a test comment', got %q", body["message"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.Comment{
			ID:         "cmt-new",
			Message:    "This is a test comment",
			AuthorID:   "user-1",
			AuthorName: "Alice",
			CreatedAt:  "2024-06-20T15:00:00Z",
		})
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.CreateComment("proj-1", "err-1", "This is a test comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "cmt-new" {
		t.Errorf("expected ID=cmt-new, got %s", result.ID)
	}
	if result.Message != "This is a test comment" {
		t.Errorf("unexpected message: %s", result.Message)
	}
	if result.AuthorName != "Alice" {
		t.Errorf("expected AuthorName=Alice, got %s", result.AuthorName)
	}
}

func TestCreateComment_EmptyMessage(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		var body map[string]string
		json.Unmarshal(bodyBytes, &body)
		if body["message"] != "" {
			t.Errorf("expected empty message, got %q", body["message"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.Comment{ID: "cmt-1", Message: ""})
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.CreateComment("proj-1", "err-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "cmt-1" {
		t.Errorf("expected ID=cmt-1, got %s", result.ID)
	}
}

func TestCreateComment_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(403, "Forbidden"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.CreateComment("proj-1", "err-1", "test")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 403 {
		t.Errorf("expected 403, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Forbidden" {
		t.Errorf("unexpected message: %s", apiErr.Message)
	}
}

func TestCreateComment_ServerError(t *testing.T) {
	server := newTestServer(t, errorHandler(500, "Internal Server Error"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.CreateComment("proj-1", "err-1", "boom")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Releases
// ===========================================================================

func TestListReleases_SinglePage(t *testing.T) {
	releases := []models.Release{
		{
			ID:            "rel-1",
			ProjectID:     "proj-1",
			Version:       "1.0.0",
			ReleaseStage:  models.ReleaseStage{Name: "production"},
			BuilderName:   "CI Bot",
			ReleaseSource: "buildkite",
			ReleaseTime:   "2024-06-01T08:00:00Z",
			TotalSessionsCount:     1000,
			UnhandledSessionsCount: 10,
			ErrorsIntroducedCount:  2,
			ErrorsSeenCount:        5,
		},
		{
			ID:           "rel-2",
			ProjectID:    "proj-1",
			Version:      "1.1.0",
			ReleaseStage: models.ReleaseStage{Name: "staging"},
			ReleaseTime:  "2024-06-10T12:00:00Z",
		},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/releases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListReleases("proj-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 releases, got %d", len(result))
	}
	if result[0].Version != "1.0.0" {
		t.Errorf("expected Version=1.0.0, got %s", result[0].Version)
	}
	if result[0].ReleaseStage.Name != "production" {
		t.Errorf("expected ReleaseStage=production, got %s", result[0].ReleaseStage.Name)
	}
	if result[0].TotalSessionsCount != 1000 {
		t.Errorf("expected TotalSessionsCount=1000, got %d", result[0].TotalSessionsCount)
	}
	if result[0].ErrorsIntroducedCount != 2 {
		t.Errorf("expected ErrorsIntroducedCount=2, got %d", result[0].ErrorsIntroducedCount)
	}
	if result[1].ReleaseStage.Name != "staging" {
		t.Errorf("expected ReleaseStage=staging, got %s", result[1].ReleaseStage.Name)
	}
}

func TestListReleases_AllPages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/projects/proj-1/releases?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Release{{ID: "rel-1", Version: "1.0.0"}})
		} else {
			json.NewEncoder(w).Encode([]models.Release{{ID: "rel-2", Version: "2.0.0"}})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListReleases("proj-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("allPages should return hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 releases, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestListReleases_APIError(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Project not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, _, err := c.ListReleases("bad-proj", false)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Stability Trend
// ===========================================================================

func TestGetStabilityTrend_Success(t *testing.T) {
	trend := models.StabilityTrend{
		ProjectID:    "proj-1",
		ReleaseStage: "production",
		TimelinePoints: []models.TimelinePoint{
			{
				BucketStart:            "2024-06-01",
				BucketEnd:              "2024-06-08",
				TotalSessionsCount:     5000,
				UnhandledSessionsCount: 50,
				UnhandledRate:          0.01,
				UsersSeen:              1000,
				UsersWithUnhandled:     10,
				UnhandledUserRate:      0.01,
			},
			{
				BucketStart:            "2024-06-08",
				BucketEnd:              "2024-06-15",
				TotalSessionsCount:     6000,
				UnhandledSessionsCount: 30,
				UnhandledRate:          0.005,
				UsersSeen:              1200,
				UsersWithUnhandled:     5,
				UnhandledUserRate:      0.0042,
			},
		},
	}
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-1/stability_trend" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("release_stage") != "production" {
			t.Errorf("expected release_stage=production, got %s", q.Get("release_stage"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trend)
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetStabilityTrend("proj-1", "production")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProjectID != "proj-1" {
		t.Errorf("expected ProjectID=proj-1, got %s", result.ProjectID)
	}
	if result.ReleaseStage != "production" {
		t.Errorf("expected ReleaseStage=production, got %s", result.ReleaseStage)
	}
	if len(result.TimelinePoints) != 2 {
		t.Fatalf("expected 2 timeline points, got %d", len(result.TimelinePoints))
	}
	tp0 := result.TimelinePoints[0]
	if tp0.TotalSessionsCount != 5000 {
		t.Errorf("expected TotalSessionsCount=5000, got %d", tp0.TotalSessionsCount)
	}
	if tp0.UnhandledSessionsCount != 50 {
		t.Errorf("expected UnhandledSessionsCount=50, got %d", tp0.UnhandledSessionsCount)
	}
	if tp0.UnhandledRate != 0.01 {
		t.Errorf("expected UnhandledRate=0.01, got %f", tp0.UnhandledRate)
	}
	if tp0.UsersSeen != 1000 {
		t.Errorf("expected UsersSeen=1000, got %d", tp0.UsersSeen)
	}
	tp1 := result.TimelinePoints[1]
	if tp1.BucketStart != "2024-06-08" {
		t.Errorf("expected BucketStart=2024-06-08, got %s", tp1.BucketStart)
	}
	if tp1.UsersWithUnhandled != 5 {
		t.Errorf("expected UsersWithUnhandled=5, got %d", tp1.UsersWithUnhandled)
	}
}

func TestGetStabilityTrend_NoReleaseStage(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("release_stage") != "" {
			t.Errorf("expected no release_stage param, got %s", q.Get("release_stage"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.StabilityTrend{
			ProjectID:      "proj-1",
			TimelinePoints: []models.TimelinePoint{},
		})
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetStabilityTrend("proj-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProjectID != "proj-1" {
		t.Errorf("expected ProjectID=proj-1, got %s", result.ProjectID)
	}
	if len(result.TimelinePoints) != 0 {
		t.Errorf("expected 0 timeline points, got %d", len(result.TimelinePoints))
	}
}

func TestGetStabilityTrend_NotFound(t *testing.T) {
	server := newTestServer(t, errorHandler(404, "Project not found"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetStabilityTrend("bad-proj", "production")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestGetStabilityTrend_ServerError(t *testing.T) {
	server := newTestServer(t, errorHandler(500, "Internal Server Error"))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, err := c.GetStabilityTrend("proj-1", "production")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Auth header verification across resource types
// ===========================================================================

func TestAuthHeaderSentOnAllRequests(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "token secret-token-123" {
			t.Errorf("expected 'token secret-token-123', got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		// Return appropriate responses based on path and method
		switch {
		case r.URL.Path == "/user/organizations":
			json.NewEncoder(w).Encode([]models.Organization{})
		case r.URL.Path == "/projects/p/stability_trend":
			json.NewEncoder(w).Encode(models.StabilityTrend{})
		case r.URL.Path == "/projects/p/trend":
			json.NewEncoder(w).Encode([]models.TrendBucket{})
		case r.URL.Path == "/projects/p/errors/e/trend":
			json.NewEncoder(w).Encode([]models.TrendBucket{})
		case r.URL.Path == "/projects/p" && r.Method == "GET":
			json.NewEncoder(w).Encode(models.Project{})
		case r.URL.Path == "/projects/p/errors/e" && r.Method == "GET":
			json.NewEncoder(w).Encode(models.BugsnagError{})
		case r.URL.Path == "/projects/p/events/e" && r.Method == "GET":
			json.NewEncoder(w).Encode(models.Event{})
		case strings.HasSuffix(r.URL.Path, "/comments") && r.Method == "POST":
			json.NewEncoder(w).Encode(models.Comment{})
		default:
			json.NewEncoder(w).Encode([]json.RawMessage{})
		}
	})
	defer server.Close()

	c := New(server.URL, "secret-token-123", 30)

	// Exercise various resource methods to verify auth header is always sent
	c.ListOrganizations(false)
	c.GetProject("p")
	c.GetError("p", "e")
	c.GetEvent("p", "e")
	c.GetProjectTrends("p", "", 0)
	c.GetErrorTrends("p", "e")
	c.GetStabilityTrend("p", "")
	c.CreateComment("p", "e", "msg")
}

// ===========================================================================
// Multi-page pagination (3 pages) to test deeper iteration
// ===========================================================================

func TestListOrganizations_ThreePages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		switch callCount {
		case 1:
			nextURL := "http://" + r.Host + "/user/organizations?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Organization{{ID: "org-1"}})
		case 2:
			nextURL := "http://" + r.Host + "/user/organizations?offset=p3"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Organization{{ID: "org-2"}})
		case 3:
			json.NewEncoder(w).Encode([]models.Organization{{ID: "org-3"}})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, _, err := c.ListOrganizations(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 orgs across 3 pages, got %d", len(result))
	}
	if callCount != 3 {
		t.Errorf("expected 3 API calls, got %d", callCount)
	}
	if result[0].ID != "org-1" || result[1].ID != "org-2" || result[2].ID != "org-3" {
		t.Errorf("unexpected order: %s, %s, %s", result[0].ID, result[1].ID, result[2].ID)
	}
}

// ===========================================================================
// Release with SourceControl sub-object
// ===========================================================================

func TestListReleases_WithSourceControl(t *testing.T) {
	releases := []models.Release{
		{
			ID:      "rel-sc",
			Version: "3.0.0",
			SourceControl: &models.SourceControl{
				Provider:   "github",
				Revision:   "abc123def456",
				Repository: "https://github.com/org/repo",
				DiffURL:    "https://github.com/org/repo/compare/v2...v3",
			},
		},
	}
	server := newTestServer(t, jsonHandler(t, 200, releases))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, _, err := c.ListReleases("proj-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 release, got %d", len(result))
	}
	if result[0].SourceControl == nil {
		t.Fatal("expected non-nil SourceControl")
	}
	sc := result[0].SourceControl
	if sc.Provider != "github" {
		t.Errorf("expected Provider=github, got %s", sc.Provider)
	}
	if sc.Revision != "abc123def456" {
		t.Errorf("expected Revision=abc123def456, got %s", sc.Revision)
	}
	if sc.Repository != "https://github.com/org/repo" {
		t.Errorf("unexpected Repository: %s", sc.Repository)
	}
}

// ===========================================================================
// Error with nested objects (Issue, Overrides, GroupingFields)
// ===========================================================================

func TestGetError_WithNestedObjects(t *testing.T) {
	bugErr := models.BugsnagError{
		ID:         "err-nested",
		ErrorClass: "SomeError",
		Severity:   "error",
		Status:     "open",
		CreatedIssue: &models.Issue{
			ID:     "issue-1",
			Number: "42",
			Type:   "jira",
			URL:    "https://jira.example.com/browse/BUG-42",
		},
		Overrides: &models.ErrorOverrides{
			Severity: "warning",
		},
		GroupingFields: &models.GroupingFields{
			ErrorClass: "SomeError",
			File:       "main.go",
			Linenum:    42,
		},
	}
	server := newTestServer(t, jsonHandler(t, 200, bugErr))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetError("proj-1", "err-nested")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CreatedIssue == nil {
		t.Fatal("expected non-nil CreatedIssue")
	}
	if result.CreatedIssue.Number != "42" {
		t.Errorf("expected Issue.Number=42, got %s", result.CreatedIssue.Number)
	}
	if result.CreatedIssue.Type != "jira" {
		t.Errorf("expected Issue.Type=jira, got %s", result.CreatedIssue.Type)
	}
	if result.Overrides == nil {
		t.Fatal("expected non-nil Overrides")
	}
	if result.Overrides.Severity != "warning" {
		t.Errorf("expected Overrides.Severity=warning, got %s", result.Overrides.Severity)
	}
	if result.GroupingFields == nil {
		t.Fatal("expected non-nil GroupingFields")
	}
	if result.GroupingFields.File != "main.go" {
		t.Errorf("expected GroupingFields.File=main.go, got %s", result.GroupingFields.File)
	}
	if result.GroupingFields.Linenum != 42 {
		t.Errorf("expected GroupingFields.Linenum=42, got %d", result.GroupingFields.Linenum)
	}
}

// ===========================================================================
// CreateComment: verify JSON body structure with special characters
// ===========================================================================

func TestCreateComment_SpecialCharactersInMessage(t *testing.T) {
	msg := `Line1\nLine2 "quoted" <html>&amp; special`
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		var body map[string]string
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if body["message"] != msg {
			t.Errorf("expected message=%q, got %q", msg, body["message"])
		}
		// Verify only "message" key is sent
		if len(body) != 1 {
			t.Errorf("expected exactly 1 key in body, got %d", len(body))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.Comment{ID: "cmt-special", Message: msg})
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.CreateComment("proj-1", "err-1", msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Message != msg {
		t.Errorf("expected message=%q, got %q", msg, result.Message)
	}
}

// ===========================================================================
// AllPages error propagation: error on the second page
// ===========================================================================

func TestListErrors_AllPages_ErrorOnSecondPage(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/projects/proj-1/errors?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.BugsnagError{{ID: "err-1"}})
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]any{
				"errors": []map[string]string{{"message": "Server crashed on page 2"}},
			})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	_, _, err := c.ListErrors(ListErrorsOptions{ProjectID: "proj-1", AllPages: true})
	if err == nil {
		t.Fatal("expected error when second page fails")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

// ===========================================================================
// Event with JSON raw message fields
// ===========================================================================

func TestGetEvent_WithRawJSONFields(t *testing.T) {
	rawJSON := `{
		"id": "evt-raw",
		"project_id": "proj-1",
		"error_id": "err-1",
		"error_class": "Error",
		"severity": "error",
		"received_at": "2024-01-01T00:00:00Z",
		"context": "main",
		"message": "something failed",
		"unhandled": false,
		"url": "https://example.com",
		"app": {"version": "1.0.0", "release_stage": "production"},
		"device": {"os": "linux", "hostname": "server-1"},
		"user": {"id": "u1", "email": "test@example.com"},
		"meta_data": {"custom_key": "custom_value"}
	}`
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(rawJSON))
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := c.GetEvent("proj-1", "evt-raw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "evt-raw" {
		t.Errorf("expected ID=evt-raw, got %s", result.ID)
	}
	// Verify raw JSON fields are captured
	if result.App == nil {
		t.Error("expected non-nil App")
	}
	if result.Device == nil {
		t.Error("expected non-nil Device")
	}
	if result.User == nil {
		t.Error("expected non-nil User")
	}
	if result.MetaData == nil {
		t.Error("expected non-nil MetaData")
	}

	// Verify the raw JSON can be parsed
	var app map[string]string
	if err := json.Unmarshal(result.App, &app); err != nil {
		t.Fatalf("failed to parse App JSON: %v", err)
	}
	if app["version"] != "1.0.0" {
		t.Errorf("expected App.version=1.0.0, got %s", app["version"])
	}
}

// ===========================================================================
// ListEvents for error with allPages
// ===========================================================================

func TestListEvents_ForError_AllPages(t *testing.T) {
	callCount := 0
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			if r.URL.Path != "/projects/proj-1/errors/err-1/events" {
				t.Errorf("unexpected path on first call: %s", r.URL.Path)
			}
			nextURL := "http://" + r.Host + "/projects/proj-1/errors/err-1/events?offset=p2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]models.Event{{ID: "evt-1", ErrorID: "err-1"}})
		} else {
			json.NewEncoder(w).Encode([]models.Event{{ID: "evt-2", ErrorID: "err-1"}})
		}
	})
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListEvents("proj-1", "err-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("allPages should return hasMore=false")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

// ===========================================================================
// Empty list responses
// ===========================================================================

func TestListOrganizations_EmptyResult(t *testing.T) {
	server := newTestServer(t, jsonHandler(t, 200, []models.Organization{}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := c.ListOrganizations(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 orgs, got %d", len(result))
	}
}

func TestListProjects_EmptyResult(t *testing.T) {
	server := newTestServer(t, jsonHandler(t, 200, []models.Project{}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, _, err := c.ListProjects("org-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 projects, got %d", len(result))
	}
}

func TestListReleases_EmptyResult(t *testing.T) {
	server := newTestServer(t, jsonHandler(t, 200, []models.Release{}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, _, err := c.ListReleases("proj-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 releases, got %d", len(result))
	}
}
