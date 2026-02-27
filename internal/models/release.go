package models

type Release struct {
	ID             string `json:"id"`
	ProjectID      string `json:"project_id"`
	Version        string `json:"app_version"`
	ReleaseStage   ReleaseStage `json:"release_stage"`
	BuilderName    string `json:"builder_name"`
	ReleaseSource  string `json:"release_source"`
	ReleaseTime    string `json:"release_time"`
	BuildLabel     string `json:"build_label"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	SourceControl  *SourceControl    `json:"source_control,omitempty"`
	TotalSessionsCount     int `json:"total_sessions_count"`
	UnhandledSessionsCount int `json:"unhandled_sessions_count"`
	ErrorsIntroducedCount  int `json:"errors_introduced_count"`
	ErrorsSeenCount        int `json:"errors_seen_count"`
}

type ReleaseStage struct {
	Name string `json:"name"`
}

type SourceControl struct {
	Provider   string `json:"provider"`
	Revision   string `json:"revision"`
	Repository string `json:"repository"`
	DiffURL    string `json:"diff_url"`
}

func (r Release) TableHeaders() []string {
	return []string{"ID", "VERSION", "RELEASE_STAGE", "SOURCE", "RELEASE_TIME"}
}

func (r Release) TableRow() []string {
	return []string{r.ID, r.Version, r.ReleaseStage.Name, r.ReleaseSource, r.ReleaseTime}
}
