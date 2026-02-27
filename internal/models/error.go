package models

type BugsnagError struct {
	ID              string `json:"id"`
	ProjectID       string `json:"project_id"`
	ErrorClass      string `json:"error_class"`
	Message         string `json:"message"`
	Context         string `json:"context"`
	Severity        string `json:"severity"`
	Status          string `json:"status"`
	Unhandled       bool   `json:"unhandled"`
	FirstSeen       string `json:"first_seen"`
	LastSeen        string `json:"last_seen"`
	EventsCount     int    `json:"events"`
	Events          int    `json:"events_count"`
	UnthrottledOccurrenceCount int `json:"unthrottled_occurrence_count"`
	URL             string `json:"url"`
	AssignedCollaboratorID string `json:"assigned_collaborator_id,omitempty"`
	CommentCount    int    `json:"comment_count"`
	CreatedIssue    *Issue `json:"created_issue,omitempty"`
	OriginalSeverity string `json:"original_severity"`
	Overrides       *ErrorOverrides `json:"overrides,omitempty"`
	MissingDsyms    []string `json:"missing_dsyms,omitempty"`
	ReleaseStages   []string `json:"release_stages,omitempty"`
	GroupingReason   string  `json:"grouping_reason"`
	GroupingFields   *GroupingFields `json:"grouping_fields,omitempty"`
	Reopen          bool    `json:"reopen_rules,omitempty"`
	FirstSeenUnfiltered string `json:"first_seen_unfiltered"`
	LastSeenUnfiltered  string `json:"last_seen_unfiltered"`
}

type Issue struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Type   string `json:"type"`
	URL    string `json:"url"`
}

type ErrorOverrides struct {
	Severity string `json:"severity,omitempty"`
}

type GroupingFields struct {
	ErrorClass string `json:"error_class,omitempty"`
	File       string `json:"file,omitempty"`
	Linenum    int    `json:"linenum,omitempty"`
}

func (e BugsnagError) TableHeaders() []string {
	return []string{"ID", "ERROR_CLASS", "SEVERITY", "STATUS", "EVENTS", "LAST_SEEN"}
}

func (e BugsnagError) TableRow() []string {
	return []string{e.ID, e.ErrorClass, e.Severity, e.Status, itoa(e.EventsCount), e.LastSeen}
}
