package models

import "testing"

// ---------------------------------------------------------------------------
// Existing tests (preserved)
// ---------------------------------------------------------------------------

func TestHelpers(t *testing.T) {
	if got := itoa(42); got != "42" {
		t.Errorf("itoa(42) = %q, want \"42\"", got)
	}
	if got := itoa(0); got != "0" {
		t.Errorf("itoa(0) = %q, want \"0\"", got)
	}
	if got := ftoa(0.5); got != "0.5000" {
		t.Errorf("ftoa(0.5) = %q, want \"0.5000\"", got)
	}
}

func TestOrganizationTable(t *testing.T) {
	o := Organization{ID: "abc", Name: "My Org", Slug: "my-org", CreatedAt: "2024-01-01"}
	headers := o.TableHeaders()
	if len(headers) != 4 {
		t.Errorf("expected 4 headers, got %d", len(headers))
	}
	row := o.TableRow()
	if row[0] != "abc" || row[1] != "My Org" {
		t.Errorf("unexpected row: %v", row)
	}
}

func TestProjectTable(t *testing.T) {
	p := Project{ID: "123", Name: "Test", Language: "go", OpenErrorCount: 5, CreatedAt: "2024-01-01"}
	row := p.TableRow()
	if row[3] != "5" {
		t.Errorf("expected open errors '5', got %s", row[3])
	}
}

func TestBugsnagErrorTable(t *testing.T) {
	e := BugsnagError{ID: "err1", ErrorClass: "TypeError", Severity: "error", Status: "open", EventsCount: 10, LastSeen: "2024-01-01"}
	headers := e.TableHeaders()
	if headers[0] != "ID" {
		t.Errorf("expected first header 'ID', got %s", headers[0])
	}
	row := e.TableRow()
	if row[4] != "10" {
		t.Errorf("expected events '10', got %s", row[4])
	}
}

func TestEventTable(t *testing.T) {
	e := Event{ID: "ev1", ErrorClass: "Error", Severity: "warning", Context: "/api", ReceivedAt: "2024-01-01"}
	row := e.TableRow()
	if row[2] != "warning" {
		t.Errorf("expected severity 'warning', got %s", row[2])
	}
}

func TestTrendBucketTable(t *testing.T) {
	b := TrendBucket{From: "2024-01-01", To: "2024-01-02", EventsCount: 100}
	row := b.TableRow()
	if row[2] != "100" {
		t.Errorf("expected events '100', got %s", row[2])
	}
}

func TestCollaboratorTable(t *testing.T) {
	c := Collaborator{ID: "c1", Name: "Alice", Email: "alice@test.com", IsAdmin: true}
	row := c.TableRow()
	if row[3] != "yes" {
		t.Errorf("expected admin 'yes', got %s", row[3])
	}

	c2 := Collaborator{IsAdmin: false}
	if c2.TableRow()[3] != "no" {
		t.Error("expected admin 'no'")
	}
}

func TestCommentTable(t *testing.T) {
	c := Comment{ID: "cm1", Message: "short msg", AuthorName: "Bob", CreatedAt: "2024-01-01"}
	row := c.TableRow()
	if row[2] != "short msg" {
		t.Errorf("expected message 'short msg', got %s", row[2])
	}

	longMsg := "This is a very long comment message that should be truncated because it exceeds sixty characters in total length"
	c2 := Comment{Message: longMsg}
	row2 := c2.TableRow()
	if len(row2[2]) != 60 {
		t.Errorf("expected truncated to 60 chars, got %d", len(row2[2]))
	}
}

func TestReleaseTable(t *testing.T) {
	r := Release{
		ID:            "r1",
		Version:       "1.0.0",
		ReleaseStage:  ReleaseStage{Name: "production"},
		ReleaseSource: "api",
		ReleaseTime:   "2024-01-01",
	}
	row := r.TableRow()
	if row[2] != "production" {
		t.Errorf("expected stage 'production', got %s", row[2])
	}
}

func TestTimelinePointTable(t *testing.T) {
	tp := TimelinePoint{
		BucketStart:            "2024-01-01",
		BucketEnd:              "2024-01-02",
		TotalSessionsCount:     100,
		UnhandledSessionsCount: 10,
		UnhandledRate:          0.1,
	}
	row := tp.TableRow()
	if row[2] != "100" {
		t.Errorf("expected sessions '100', got %s", row[2])
	}
	if row[4] != "0.1000" {
		t.Errorf("expected rate '0.1000', got %s", row[4])
	}
}

// ---------------------------------------------------------------------------
// New tests: TableHeaders() coverage for all models
// ---------------------------------------------------------------------------

// -- Collaborator.TableHeaders ------------------------------------------------

func TestCollaboratorTableHeaders(t *testing.T) {
	c := Collaborator{}
	headers := c.TableHeaders()

	expected := []string{"ID", "NAME", "EMAIL", "IS_ADMIN"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// -- Comment.TableHeaders -----------------------------------------------------

func TestCommentTableHeaders(t *testing.T) {
	c := Comment{}
	headers := c.TableHeaders()

	expected := []string{"ID", "AUTHOR", "MESSAGE", "CREATED_AT"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// -- Event.TableHeaders -------------------------------------------------------

func TestEventTableHeaders(t *testing.T) {
	e := Event{}
	headers := e.TableHeaders()

	expected := []string{"ID", "ERROR_CLASS", "SEVERITY", "CONTEXT", "RECEIVED_AT"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// -- Project.TableHeaders -----------------------------------------------------

func TestProjectTableHeaders(t *testing.T) {
	p := Project{}
	headers := p.TableHeaders()

	expected := []string{"ID", "NAME", "LANGUAGE", "OPEN_ERRORS", "CREATED_AT"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// -- Release.TableHeaders -----------------------------------------------------

func TestReleaseTableHeaders(t *testing.T) {
	r := Release{}
	headers := r.TableHeaders()

	expected := []string{"ID", "VERSION", "RELEASE_STAGE", "SOURCE", "RELEASE_TIME"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// -- TimelinePoint.TableHeaders -----------------------------------------------

func TestTimelinePointTableHeaders(t *testing.T) {
	tp := TimelinePoint{}
	headers := tp.TableHeaders()

	expected := []string{"BUCKET_START", "BUCKET_END", "SESSIONS", "UNHANDLED", "UNHANDLED_RATE"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// -- TrendBucket.TableHeaders -------------------------------------------------

func TestTrendBucketTableHeaders(t *testing.T) {
	tb := TrendBucket{}
	headers := tb.TableHeaders()

	expected := []string{"FROM", "TO", "EVENTS_COUNT"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// -- BugsnagError.TableHeaders ------------------------------------------------

func TestBugsnagErrorTableHeaders(t *testing.T) {
	e := BugsnagError{}
	headers := e.TableHeaders()

	expected := []string{"ID", "ERROR_CLASS", "SEVERITY", "STATUS", "EVENTS", "LAST_SEEN"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// -- Organization.TableHeaders ------------------------------------------------

func TestOrganizationTableHeaders(t *testing.T) {
	o := Organization{}
	headers := o.TableHeaders()

	expected := []string{"ID", "NAME", "SLUG", "CREATED_AT"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, headers[i])
		}
	}
}

// ---------------------------------------------------------------------------
// New tests: TableRow() edge cases
// ---------------------------------------------------------------------------

// -- Collaborator with no admin -----------------------------------------------

func TestCollaboratorTableRow_NotAdmin(t *testing.T) {
	c := Collaborator{ID: "c2", Name: "Bob", Email: "bob@test.com", IsAdmin: false}
	row := c.TableRow()
	if row[0] != "c2" {
		t.Errorf("expected id 'c2', got %q", row[0])
	}
	if row[1] != "Bob" {
		t.Errorf("expected name 'Bob', got %q", row[1])
	}
	if row[2] != "bob@test.com" {
		t.Errorf("expected email 'bob@test.com', got %q", row[2])
	}
	if row[3] != "no" {
		t.Errorf("expected admin 'no', got %q", row[3])
	}
}

// -- Comment with exactly 60 characters (boundary) ---------------------------

func TestCommentTableRow_Exactly60Chars(t *testing.T) {
	// 60 characters exactly
	msg := "123456789012345678901234567890123456789012345678901234567890"
	if len(msg) != 60 {
		t.Fatalf("test setup: expected 60 chars, got %d", len(msg))
	}
	c := Comment{ID: "cm", Message: msg}
	row := c.TableRow()
	if row[2] != msg {
		t.Errorf("expected message not to be truncated at exactly 60 chars")
	}
}

func TestCommentTableRow_61Chars(t *testing.T) {
	// 61 characters -- should be truncated
	msg := "1234567890123456789012345678901234567890123456789012345678901"
	if len(msg) != 61 {
		t.Fatalf("test setup: expected 61 chars, got %d", len(msg))
	}
	c := Comment{ID: "cm", Message: msg}
	row := c.TableRow()
	if len(row[2]) != 60 {
		t.Errorf("expected truncated to 60 chars, got %d", len(row[2]))
	}
	if row[2][57:] != "..." {
		t.Errorf("expected trailing '...', got %q", row[2][57:])
	}
}

// -- Empty fields produce empty strings, not panics ---------------------------

func TestCollaboratorTableRow_EmptyFields(t *testing.T) {
	c := Collaborator{}
	row := c.TableRow()
	if len(row) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(row))
	}
	if row[0] != "" {
		t.Errorf("expected empty id, got %q", row[0])
	}
}

func TestCommentTableRow_EmptyFields(t *testing.T) {
	c := Comment{}
	row := c.TableRow()
	if len(row) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(row))
	}
}

func TestEventTableRow_EmptyFields(t *testing.T) {
	e := Event{}
	row := e.TableRow()
	if len(row) != 5 {
		t.Fatalf("expected 5 columns, got %d", len(row))
	}
}

func TestProjectTableRow_EmptyFields(t *testing.T) {
	p := Project{}
	row := p.TableRow()
	if len(row) != 5 {
		t.Fatalf("expected 5 columns, got %d", len(row))
	}
	if row[3] != "0" {
		t.Errorf("expected open_error_count '0', got %q", row[3])
	}
}

func TestReleaseTableRow_EmptyFields(t *testing.T) {
	r := Release{}
	row := r.TableRow()
	if len(row) != 5 {
		t.Fatalf("expected 5 columns, got %d", len(row))
	}
}

func TestTimelinePointTableRow_EmptyFields(t *testing.T) {
	tp := TimelinePoint{}
	row := tp.TableRow()
	if len(row) != 5 {
		t.Fatalf("expected 5 columns, got %d", len(row))
	}
	if row[2] != "0" {
		t.Errorf("expected sessions '0', got %q", row[2])
	}
	if row[4] != "0.0000" {
		t.Errorf("expected rate '0.0000', got %q", row[4])
	}
}

func TestTrendBucketTableRow_EmptyFields(t *testing.T) {
	tb := TrendBucket{}
	row := tb.TableRow()
	if len(row) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(row))
	}
	if row[2] != "0" {
		t.Errorf("expected events '0', got %q", row[2])
	}
}

func TestBugsnagErrorTableRow_EmptyFields(t *testing.T) {
	e := BugsnagError{}
	row := e.TableRow()
	if len(row) != 6 {
		t.Fatalf("expected 6 columns, got %d", len(row))
	}
	if row[4] != "0" {
		t.Errorf("expected events '0', got %q", row[4])
	}
}

func TestOrganizationTableRow_EmptyFields(t *testing.T) {
	o := Organization{}
	row := o.TableRow()
	if len(row) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(row))
	}
}

// ---------------------------------------------------------------------------
// Helpers edge cases
// ---------------------------------------------------------------------------

func TestItoa_Negative(t *testing.T) {
	if got := itoa(-1); got != "-1" {
		t.Errorf("itoa(-1) = %q, want \"-1\"", got)
	}
}

func TestItoa_Large(t *testing.T) {
	if got := itoa(1000000); got != "1000000" {
		t.Errorf("itoa(1000000) = %q, want \"1000000\"", got)
	}
}

func TestFtoa_Zero(t *testing.T) {
	if got := ftoa(0.0); got != "0.0000" {
		t.Errorf("ftoa(0.0) = %q, want \"0.0000\"", got)
	}
}

func TestFtoa_Negative(t *testing.T) {
	if got := ftoa(-0.123); got != "-0.1230" {
		t.Errorf("ftoa(-0.123) = %q, want \"-0.1230\"", got)
	}
}

func TestFtoa_LargeValue(t *testing.T) {
	got := ftoa(99.9999)
	if got != "99.9999" {
		t.Errorf("ftoa(99.9999) = %q, want \"99.9999\"", got)
	}
}

// ---------------------------------------------------------------------------
// Full round-trip: headers count matches row count
// ---------------------------------------------------------------------------

func TestHeadersMatchRowCount_Collaborator(t *testing.T) {
	c := Collaborator{ID: "1", Name: "N", Email: "e", IsAdmin: true}
	if len(c.TableHeaders()) != len(c.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(c.TableHeaders()), len(c.TableRow()))
	}
}

func TestHeadersMatchRowCount_Comment(t *testing.T) {
	c := Comment{ID: "1", AuthorName: "A", Message: "m", CreatedAt: "t"}
	if len(c.TableHeaders()) != len(c.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(c.TableHeaders()), len(c.TableRow()))
	}
}

func TestHeadersMatchRowCount_Event(t *testing.T) {
	e := Event{ID: "1", ErrorClass: "E", Severity: "s", Context: "c", ReceivedAt: "r"}
	if len(e.TableHeaders()) != len(e.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(e.TableHeaders()), len(e.TableRow()))
	}
}

func TestHeadersMatchRowCount_Project(t *testing.T) {
	p := Project{ID: "1", Name: "N", Language: "L", OpenErrorCount: 0, CreatedAt: "c"}
	if len(p.TableHeaders()) != len(p.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(p.TableHeaders()), len(p.TableRow()))
	}
}

func TestHeadersMatchRowCount_Release(t *testing.T) {
	r := Release{ID: "1", Version: "v", ReleaseStage: ReleaseStage{Name: "s"}, ReleaseSource: "src", ReleaseTime: "t"}
	if len(r.TableHeaders()) != len(r.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(r.TableHeaders()), len(r.TableRow()))
	}
}

func TestHeadersMatchRowCount_TimelinePoint(t *testing.T) {
	tp := TimelinePoint{BucketStart: "s", BucketEnd: "e", TotalSessionsCount: 1, UnhandledSessionsCount: 0, UnhandledRate: 0.0}
	if len(tp.TableHeaders()) != len(tp.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(tp.TableHeaders()), len(tp.TableRow()))
	}
}

func TestHeadersMatchRowCount_TrendBucket(t *testing.T) {
	tb := TrendBucket{From: "f", To: "t", EventsCount: 1}
	if len(tb.TableHeaders()) != len(tb.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(tb.TableHeaders()), len(tb.TableRow()))
	}
}

func TestHeadersMatchRowCount_BugsnagError(t *testing.T) {
	e := BugsnagError{ID: "1", ErrorClass: "E", Severity: "s", Status: "o", EventsCount: 0, LastSeen: "l"}
	if len(e.TableHeaders()) != len(e.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(e.TableHeaders()), len(e.TableRow()))
	}
}

func TestHeadersMatchRowCount_Organization(t *testing.T) {
	o := Organization{ID: "1", Name: "N", Slug: "s", CreatedAt: "c"}
	if len(o.TableHeaders()) != len(o.TableRow()) {
		t.Errorf("headers count %d != row count %d", len(o.TableHeaders()), len(o.TableRow()))
	}
}

// -- Release with SourceControl -----------------------------------------------

func TestReleaseTableRow_WithSourceControl(t *testing.T) {
	r := Release{
		ID:            "r2",
		Version:       "2.0.0",
		ReleaseStage:  ReleaseStage{Name: "staging"},
		ReleaseSource: "dashboard",
		ReleaseTime:   "2024-06-15",
		SourceControl: &SourceControl{
			Provider:   "github",
			Revision:   "abc123",
			Repository: "org/repo",
		},
	}
	row := r.TableRow()
	if row[1] != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %q", row[1])
	}
	if row[3] != "dashboard" {
		t.Errorf("expected source 'dashboard', got %q", row[3])
	}
}

// -- TimelinePoint with all fields populated ----------------------------------

func TestTimelinePointTableRow_AllFields(t *testing.T) {
	tp := TimelinePoint{
		BucketStart:            "2024-03-01",
		BucketEnd:              "2024-03-02",
		TotalSessionsCount:     5000,
		UnhandledSessionsCount: 250,
		UnhandledRate:          0.05,
		UsersSeen:              1000,
		UsersWithUnhandled:     50,
		UnhandledUserRate:      0.05,
	}
	row := tp.TableRow()
	if row[0] != "2024-03-01" {
		t.Errorf("expected bucket_start '2024-03-01', got %q", row[0])
	}
	if row[1] != "2024-03-02" {
		t.Errorf("expected bucket_end '2024-03-02', got %q", row[1])
	}
	if row[2] != "5000" {
		t.Errorf("expected sessions '5000', got %q", row[2])
	}
	if row[3] != "250" {
		t.Errorf("expected unhandled '250', got %q", row[3])
	}
	if row[4] != "0.0500" {
		t.Errorf("expected rate '0.0500', got %q", row[4])
	}
}

// -- BugsnagError with all fields populated -----------------------------------

func TestBugsnagErrorTableRow_AllFields(t *testing.T) {
	e := BugsnagError{
		ID:         "err99",
		ErrorClass: "ReferenceError",
		Severity:   "warning",
		Status:     "fixed",
		EventsCount: 999,
		LastSeen:   "2024-12-31",
	}
	row := e.TableRow()
	if row[0] != "err99" {
		t.Errorf("expected id 'err99', got %q", row[0])
	}
	if row[1] != "ReferenceError" {
		t.Errorf("expected class 'ReferenceError', got %q", row[1])
	}
	if row[2] != "warning" {
		t.Errorf("expected severity 'warning', got %q", row[2])
	}
	if row[3] != "fixed" {
		t.Errorf("expected status 'fixed', got %q", row[3])
	}
	if row[4] != "999" {
		t.Errorf("expected events '999', got %q", row[4])
	}
	if row[5] != "2024-12-31" {
		t.Errorf("expected last_seen '2024-12-31', got %q", row[5])
	}
}
