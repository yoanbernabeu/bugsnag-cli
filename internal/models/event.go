package models

import "encoding/json"

type Event struct {
	ID             string          `json:"id"`
	ProjectID      string          `json:"project_id"`
	ErrorID        string          `json:"error_id"`
	ReceivedAt     string          `json:"received_at"`
	Severity       string          `json:"severity"`
	Unhandled      bool            `json:"unhandled"`
	Context        string          `json:"context"`
	ErrorClass     string          `json:"error_class"`
	Message        string          `json:"message"`
	App            json.RawMessage `json:"app,omitempty"`
	Device         json.RawMessage `json:"device,omitempty"`
	User           json.RawMessage `json:"user,omitempty"`
	Breadcrumbs    json.RawMessage `json:"breadcrumbs,omitempty"`
	Exceptions     json.RawMessage `json:"exceptions,omitempty"`
	Threads        json.RawMessage `json:"threads,omitempty"`
	MetaData       json.RawMessage `json:"meta_data,omitempty"`
	Request        json.RawMessage `json:"request,omitempty"`
	URL            string          `json:"url"`
}

func (e Event) TableHeaders() []string {
	return []string{"ID", "ERROR_CLASS", "SEVERITY", "CONTEXT", "RECEIVED_AT"}
}

func (e Event) TableRow() []string {
	return []string{e.ID, e.ErrorClass, e.Severity, e.Context, e.ReceivedAt}
}
