package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	c := New("https://api.bugsnag.com", "my-token", 50)
	if c.BaseURL != "https://api.bugsnag.com" {
		t.Errorf("unexpected BaseURL: %s", c.BaseURL)
	}
	if c.Token != "my-token" {
		t.Errorf("unexpected Token: %s", c.Token)
	}
	if c.PerPage != 50 {
		t.Errorf("unexpected PerPage: %d", c.PerPage)
	}
	if c.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func TestNewRequest(t *testing.T) {
	c := New("https://api.bugsnag.com", "test-token", 30)
	req, err := c.newRequest("GET", "/user/organizations", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("expected GET, got %s", req.Method)
	}
	if req.Header.Get("Authorization") != "token test-token" {
		t.Errorf("unexpected auth header: %s", req.Header.Get("Authorization"))
	}
	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("unexpected accept header: %s", req.Header.Get("Accept"))
	}
	if req.URL.Query().Get("per_page") != "30" {
		t.Errorf("expected per_page=30, got %s", req.URL.Query().Get("per_page"))
	}
}

func TestNewRequestWithParams(t *testing.T) {
	c := New("https://api.bugsnag.com", "test-token", 25)
	params := map[string][]string{
		"status":   {"open"},
		"severity": {"error"},
	}
	req, err := c.newRequest("GET", "/projects/123/errors", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := req.URL.Query()
	if q.Get("status") != "open" {
		t.Errorf("expected status=open, got %s", q.Get("status"))
	}
	if q.Get("severity") != "error" {
		t.Errorf("expected severity=error, got %s", q.Get("severity"))
	}
	if q.Get("per_page") != "25" {
		t.Errorf("expected per_page=25, got %s", q.Get("per_page"))
	}
}

func TestDoSuccess(t *testing.T) {
	expected := map[string]string{"id": "123", "name": "test"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	req, _ := c.newRequest("GET", "/test", nil)

	var result map[string]string
	_, err := c.do(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("expected id=123, got %s", result["id"])
	}
}

func TestDoAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{{"message": "Bad Credentials"}},
		})
	}))
	defer server.Close()

	c := New(server.URL, "bad-token", 30)
	req, _ := c.newRequest("GET", "/test", nil)

	var result map[string]string
	_, err := c.do(req, &result)
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
	if apiErr.Message != "Bad Credentials" {
		t.Errorf("unexpected message: %s", apiErr.Message)
	}
}

func TestDoAPIErrorPlainText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	req, _ := c.newRequest("GET", "/test", nil)

	_, err := c.do(req, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
}

func TestDoNilTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	c := New(server.URL, "test-token", 30)
	req, _ := c.newRequest("GET", "/test", nil)

	_, err := c.do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAPIErrorString(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "Not Found"}
	if err.Error() != "API error (404): Not Found" {
		t.Errorf("unexpected error string: %s", err.Error())
	}

	err2 := &APIError{StatusCode: 500}
	if err2.Error() != "API error (500)" {
		t.Errorf("unexpected error string: %s", err2.Error())
	}
}

func TestNewRequestInvalidURL(t *testing.T) {
	c := New("://bad-url", "tok", 30)
	_, err := c.newRequest("GET", "/test", nil)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestNewRequestWithBodyInvalidURL(t *testing.T) {
	c := New("://bad-url", "tok", 30)
	_, err := c.newRequestWithBody("POST", "/test", nil)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestDoInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	c := New(server.URL, "tok", 30)
	req, _ := c.newRequest("GET", "/test", nil)

	var result map[string]string
	_, err := c.do(req, &result)
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decoding response") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewRequestPerPageNotOverridden(t *testing.T) {
	c := New("https://api.test.com", "tok", 30)
	params := map[string][]string{"per_page": {"50"}}
	req, err := c.newRequest("GET", "/test", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.URL.Query().Get("per_page") != "50" {
		t.Errorf("per_page should not be overridden, got %s", req.URL.Query().Get("per_page"))
	}
}

func TestFetchPageAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{{"message": "Forbidden"}},
		})
	}))
	defer server.Close()

	c := New(server.URL, "tok", 30)
	_, _, err := FetchSinglePage[testItem](c, "/test", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("expected 403, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Forbidden" {
		t.Errorf("expected 'Forbidden', got %s", apiErr.Message)
	}
}

func TestFetchPageInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json array"))
	}))
	defer server.Close()

	c := New(server.URL, "tok", 30)
	_, _, err := FetchSinglePage[testItem](c, "/test", nil)
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decoding response") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCollectAllPagesErrorOnSecondPage(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/page2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			json.NewEncoder(w).Encode([]testItem{{ID: "1", Name: "first"}})
		} else {
			w.WriteHeader(500)
			w.Write([]byte(`{"errors":[{"message":"server down"}]}`))
		}
	}))
	defer server.Close()

	c := New(server.URL, "tok", 30)
	_, err := CollectAllPages[testItem](c, "/test", nil)
	if err == nil {
		t.Fatal("expected error on second page")
	}
}

func TestFetchSinglePageBuildError(t *testing.T) {
	c := New("://bad", "tok", 30)
	_, _, err := FetchSinglePage[testItem](c, "/test", nil)
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestCollectAllPagesBuildError(t *testing.T) {
	c := New("://bad", "tok", 30)
	_, err := CollectAllPages[testItem](c, "/test", nil)
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}
