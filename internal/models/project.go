package models

type Project struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	APIKey          string   `json:"api_key"`
	Type            string   `json:"type"`
	IsFullView      bool     `json:"is_full_view"`
	ReleaseStages   []string `json:"release_stages,omitempty"`
	Language        string   `json:"language"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	URL             string   `json:"url"`
	HTMLURL         string   `json:"html_url"`
	OpenErrorCount  int      `json:"open_error_count"`
	ForReview       int      `json:"for_review"`
	CollaboratorCount int    `json:"collaborator_count"`
	GlobalGrouping  []string `json:"global_grouping,omitempty"`
	LocationGrouping []string `json:"location_grouping,omitempty"`
	DiscardedAppVersions []string `json:"discarded_app_versions,omitempty"`
	DiscardedErrors      []string `json:"discarded_errors,omitempty"`
	URLWhitelist         []string `json:"url_whitelist,omitempty"`
	IgnoreOldBrowsers    bool     `json:"ignore_old_browsers"`
	IgnoredBrowserVersions map[string]int `json:"ignored_browser_versions,omitempty"`
	ResolveOnDeploy bool `json:"resolve_on_deploy"`
	CustomEventFieldsUsed int `json:"custom_event_fields_used"`
}

func (p Project) TableHeaders() []string {
	return []string{"ID", "NAME", "LANGUAGE", "OPEN_ERRORS", "CREATED_AT"}
}

func (p Project) TableRow() []string {
	return []string{p.ID, p.Name, p.Language, itoa(p.OpenErrorCount), p.CreatedAt}
}
