package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseLinkHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "valid next link",
			header: `<https://api.bugsnag.com/projects?offset=abc123&per_page=30>; rel="next"`,
			want:   "https://api.bugsnag.com/projects?offset=abc123&per_page=30",
		},
		{
			name:   "multiple links",
			header: `<https://api.bugsnag.com/projects?offset=abc>; rel="next", <https://api.bugsnag.com/projects?offset=xyz>; rel="last"`,
			want:   "https://api.bugsnag.com/projects?offset=abc",
		},
		{
			name:   "no next link",
			header: `<https://api.bugsnag.com/projects>; rel="last"`,
			want:   "",
		},
		{
			name:   "empty header",
			header: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLinkHeader(tt.header)
			if got != tt.want {
				t.Errorf("parseLinkHeader(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestToURLValues(t *testing.T) {
	params := map[string]string{
		"status":   "open",
		"severity": "error",
		"empty":    "",
	}
	values := toURLValues(params)

	if v := values["status"]; len(v) != 1 || v[0] != "open" {
		t.Errorf("expected status=open, got %v", v)
	}
	if v := values["severity"]; len(v) != 1 || v[0] != "error" {
		t.Errorf("expected severity=error, got %v", v)
	}
	if _, ok := values["empty"]; ok {
		t.Error("empty string should be excluded")
	}
}

func TestToURLValuesNil(t *testing.T) {
	values := toURLValues(nil)
	if values != nil {
		t.Errorf("expected nil, got %v", values)
	}
}

type testItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestFetchSinglePage(t *testing.T) {
	items := []testItem{
		{ID: "1", Name: "first"},
		{ID: "2", Name: "second"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := FetchSinglePage[testItem](c, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0].ID != "1" || result[1].Name != "second" {
		t.Errorf("unexpected items: %+v", result)
	}
}

func TestFetchSinglePageWithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", `<https://api.bugsnag.com/next?offset=abc>; rel="next"`)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]testItem{{ID: "1", Name: "first"}})
	}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, hasMore, err := FetchSinglePage[testItem](c, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true")
	}
	if len(result) != 1 {
		t.Errorf("expected 1 item, got %d", len(result))
	}
}

func TestCollectAllPages(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/page2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]testItem{{ID: "1", Name: "first"}})
		} else {
			json.NewEncoder(w).Encode([]testItem{{ID: "2", Name: "second"}})
		}
	}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	result, err := CollectAllPages[testItem](c, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}
