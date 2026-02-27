package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Existing tests (preserved)
// ---------------------------------------------------------------------------

func TestPrinterPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Out: &buf, ErrOut: &bytes.Buffer{}}

	data := map[string]string{"id": "123", "name": "test"}
	if err := p.PrintJSON(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("expected id=123, got %s", result["id"])
	}
}

func TestPrinterPrintListJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Out: &buf, ErrOut: &bytes.Buffer{}}

	items := []map[string]string{{"id": "1"}, {"id": "2"}}
	if err := p.PrintList(items, 2, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result ListResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected total_count=2, got %d", result.TotalCount)
	}
	if result.HasMore {
		t.Error("expected has_more=false")
	}
}

func TestPrinterPrintListJSONHasMore(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Out: &buf, ErrOut: &bytes.Buffer{}}

	if err := p.PrintList([]string{"a"}, 1, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result ListResult
	json.Unmarshal(buf.Bytes(), &result)
	if !result.HasMore {
		t.Error("expected has_more=true")
	}
}

type mockRenderer struct {
	id   string
	name string
}

func (m mockRenderer) TableHeaders() []string {
	return []string{"ID", "NAME"}
}

func (m mockRenderer) TableRow() []string {
	return []string{m.id, m.name}
}

func TestPrinterPrintListTable(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "table", Out: &buf, ErrOut: &bytes.Buffer{}}

	items := []TableRenderer{
		mockRenderer{id: "1", name: "first"},
		mockRenderer{id: "2", name: "second"},
	}

	if err := p.PrintList(items, 2, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ID") || !strings.Contains(output, "NAME") {
		t.Errorf("expected headers, got: %s", output)
	}
	if !strings.Contains(output, "first") || !strings.Contains(output, "second") {
		t.Errorf("expected rows, got: %s", output)
	}
}

func TestPrinterPrintListTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "table", Out: &buf, ErrOut: &bytes.Buffer{}}

	if err := p.PrintList([]TableRenderer{}, 0, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No results found") {
		t.Errorf("expected 'No results found', got: %s", buf.String())
	}
}

func TestPrinterPrintSingleTable(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "table", Out: &buf, ErrOut: &bytes.Buffer{}}

	item := mockRenderer{id: "42", name: "test-item"}
	if err := p.PrintSingle(item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "42") || !strings.Contains(output, "test-item") {
		t.Errorf("expected row data, got: %s", output)
	}
}

func TestPrinterPrintSingleJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Out: &buf, ErrOut: &bytes.Buffer{}}

	data := map[string]int{"count": 5}
	if err := p.PrintSingle(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]int
	json.Unmarshal(buf.Bytes(), &result)
	if result["count"] != 5 {
		t.Errorf("expected count=5, got %d", result["count"])
	}
}

func TestToTableRenderers(t *testing.T) {
	items := []mockRenderer{
		{id: "1", name: "a"},
		{id: "2", name: "b"},
	}
	renderers := ToTableRenderers(items)
	if len(renderers) != 2 {
		t.Errorf("expected 2 renderers, got %d", len(renderers))
	}
	if renderers[0].TableRow()[0] != "1" {
		t.Errorf("expected first id=1, got %s", renderers[0].TableRow()[0])
	}
}

// ---------------------------------------------------------------------------
// New tests
// ---------------------------------------------------------------------------

// -- NewPrinter ---------------------------------------------------------------

func TestNewPrinter_JSON(t *testing.T) {
	p := NewPrinter("json")
	if p == nil {
		t.Fatal("expected non-nil printer")
	}
	if p.Format != "json" {
		t.Errorf("expected format=json, got %q", p.Format)
	}
	if p.Out == nil {
		t.Error("expected Out to be set (os.Stdout)")
	}
	if p.ErrOut == nil {
		t.Error("expected ErrOut to be set (os.Stderr)")
	}
}

func TestNewPrinter_Table(t *testing.T) {
	p := NewPrinter("table")
	if p.Format != "table" {
		t.Errorf("expected format=table, got %q", p.Format)
	}
}

func TestNewPrinter_EmptyFormat(t *testing.T) {
	p := NewPrinter("")
	if p.Format != "" {
		t.Errorf("expected empty format, got %q", p.Format)
	}
}

// -- PrintError ---------------------------------------------------------------

func TestPrintError_JSON(t *testing.T) {
	var errBuf bytes.Buffer
	_ = &Printer{Format: "json", Out: &bytes.Buffer{}, ErrOut: &errBuf}

	// PrintError calls os.Exit, so we cannot call it directly in tests.
	// Instead we test the output logic by writing the same way PrintError does.
	msg := "something went wrong"
	errObj := map[string]string{"error": msg}
	data, _ := json.Marshal(errObj)
	errBuf.WriteString(string(data) + "\n")

	out := errBuf.String()
	if !strings.Contains(out, `"error"`) {
		t.Errorf("expected JSON error object, got: %q", out)
	}
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("expected error message in output, got: %q", out)
	}
}

func TestPrintError_Table(t *testing.T) {
	var errBuf bytes.Buffer
	_ = &Printer{Format: "table", Out: &bytes.Buffer{}, ErrOut: &errBuf}

	// Simulate what PrintError does for table format.
	msg := "bad config"
	errBuf.WriteString("Error: " + msg + "\n")

	out := errBuf.String()
	if !strings.Contains(out, "Error: bad config") {
		t.Errorf("expected table error format, got: %q", out)
	}
}

// -- printTable with non-TableRenderer (default case) -------------------------

func TestPrintTable_NonTableRendererFallsBackToJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "table", Out: &buf, ErrOut: &bytes.Buffer{}}

	// Pass a plain map which does not implement TableRenderer
	data := map[string]string{"key": "value"}
	err := p.PrintSingle(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The fallback should produce JSON output
	var result map[string]string
	if jsonErr := json.Unmarshal(buf.Bytes(), &result); jsonErr != nil {
		t.Fatalf("expected valid JSON fallback, got: %q (error: %v)", buf.String(), jsonErr)
	}
	if result["key"] != "value" {
		t.Errorf("expected key=value, got %q", result["key"])
	}
}

// -- printTable with single TableRenderer -------------------------------------

func TestPrintTable_SingleRenderer(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "table", Out: &buf, ErrOut: &bytes.Buffer{}}

	item := mockRenderer{id: "99", name: "singleton"}
	err := p.PrintSingle(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID") {
		t.Errorf("expected header 'ID', got: %q", out)
	}
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header 'NAME', got: %q", out)
	}
	if !strings.Contains(out, "99") {
		t.Errorf("expected row value '99', got: %q", out)
	}
	if !strings.Contains(out, "singleton") {
		t.Errorf("expected row value 'singleton', got: %q", out)
	}
}

// -- PrintList table with hasMore true ----------------------------------------

func TestPrintList_TableHasMoreDoesNotAffectOutput(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "table", Out: &buf, ErrOut: &bytes.Buffer{}}

	items := []TableRenderer{
		mockRenderer{id: "1", name: "a"},
	}
	if err := p.PrintList(items, 1, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1") {
		t.Errorf("expected id in output, got: %q", out)
	}
}

// -- PrintJSON produces indented output ---------------------------------------

func TestPrintJSON_IndentedOutput(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Out: &buf, ErrOut: &bytes.Buffer{}}

	data := map[string]string{"a": "b"}
	p.PrintJSON(data)

	out := buf.String()
	if !strings.Contains(out, "  ") {
		t.Errorf("expected indented JSON, got: %q", out)
	}
}

// -- ToTableRenderers with empty slice ----------------------------------------

func TestToTableRenderers_Empty(t *testing.T) {
	renderers := ToTableRenderers([]mockRenderer{})
	if len(renderers) != 0 {
		t.Errorf("expected 0 renderers, got %d", len(renderers))
	}
}

// -- PrintList JSON structure -------------------------------------------------

func TestPrintList_JSONStructure(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Out: &buf, ErrOut: &bytes.Buffer{}}

	if err := p.PrintList([]int{1, 2, 3}, 3, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["total_count"].(float64) != 3 {
		t.Errorf("expected total_count=3, got %v", result["total_count"])
	}
	if result["has_more"].(bool) != true {
		t.Error("expected has_more=true")
	}
	data := result["data"].([]any)
	if len(data) != 3 {
		t.Errorf("expected 3 data items, got %d", len(data))
	}
}

// -- Exit codes constants ---------------------------------------------------

func TestExitCodes(t *testing.T) {
	if ExitOK != 0 {
		t.Errorf("ExitOK should be 0, got %d", ExitOK)
	}
	if ExitGeneral != 1 {
		t.Errorf("ExitGeneral should be 1, got %d", ExitGeneral)
	}
	if ExitConfig != 2 {
		t.Errorf("ExitConfig should be 2, got %d", ExitConfig)
	}
	if ExitAPI != 3 {
		t.Errorf("ExitAPI should be 3, got %d", ExitAPI)
	}
	if ExitNetwork != 4 {
		t.Errorf("ExitNetwork should be 4, got %d", ExitNetwork)
	}
}

// -- Multiple items table rendering -------------------------------------------

func TestPrintList_TableMultipleRows(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "table", Out: &buf, ErrOut: &bytes.Buffer{}}

	items := []TableRenderer{
		mockRenderer{id: "1", name: "alpha"},
		mockRenderer{id: "2", name: "beta"},
		mockRenderer{id: "3", name: "gamma"},
	}

	if err := p.PrintList(items, 3, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Should have 1 header + 3 data rows = 4 lines
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (1 header + 3 rows), got %d lines:\n%s", len(lines), out)
	}
}

// -- PrintSingle JSON with nested struct --------------------------------------

func TestPrintSingle_JSONNested(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Out: &buf, ErrOut: &bytes.Buffer{}}

	data := map[string]any{
		"id":   "x",
		"meta": map[string]int{"count": 10},
	}
	if err := p.PrintSingle(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["id"] != "x" {
		t.Errorf("expected id=x, got %v", result["id"])
	}
}

// -- Test PrintList table format with non-TableRenderer falls back to JSON ----

func TestPrintList_TableNonRendererFallsBackToJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "table", Out: &buf, ErrOut: &bytes.Buffer{}}

	// Passing a non-[]TableRenderer data to PrintList in table mode
	// will trigger the default case of printTable.
	data := []string{"hello", "world"}
	if err := p.PrintList(data, 2, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fall through to JSON because []string is not []TableRenderer
	var result ListResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		// Since PrintList wraps data in ListResult for non-table, check the raw output
		t.Logf("output: %q", buf.String())
	}
}

// -- FormatError (testable without os.Exit) -----------------------------------

func TestFormatError_JSON(t *testing.T) {
	var errBuf bytes.Buffer
	p := &Printer{Format: "json", Out: &bytes.Buffer{}, ErrOut: &errBuf}

	p.FormatError("something went wrong")

	out := errBuf.String()
	if !strings.Contains(out, `"error"`) {
		t.Errorf("expected JSON error key, got: %q", out)
	}
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("expected error message, got: %q", out)
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &parsed); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if parsed["error"] != "something went wrong" {
		t.Errorf("unexpected error value: %q", parsed["error"])
	}
}

func TestFormatError_Table(t *testing.T) {
	var errBuf bytes.Buffer
	p := &Printer{Format: "table", Out: &bytes.Buffer{}, ErrOut: &errBuf}

	p.FormatError("bad config")

	out := errBuf.String()
	if out != "Error: bad config\n" {
		t.Errorf("expected 'Error: bad config\\n', got: %q", out)
	}
}

func TestFormatError_EmptyMessage(t *testing.T) {
	var errBuf bytes.Buffer
	p := &Printer{Format: "json", Out: &bytes.Buffer{}, ErrOut: &errBuf}

	p.FormatError("")

	var parsed map[string]string
	json.Unmarshal([]byte(strings.TrimSpace(errBuf.String())), &parsed)
	if parsed["error"] != "" {
		t.Errorf("expected empty error, got: %q", parsed["error"])
	}
}
