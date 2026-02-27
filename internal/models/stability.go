package models

type StabilityTrend struct {
	ProjectID        string          `json:"project_id"`
	ReleaseStage     string          `json:"release_stage_name"`
	TimelinePoints   []TimelinePoint `json:"timeline_points"`
}

type TimelinePoint struct {
	BucketStart            string  `json:"bucket_start"`
	BucketEnd              string  `json:"bucket_end"`
	TotalSessionsCount     int     `json:"total_sessions_count"`
	UnhandledSessionsCount int     `json:"unhandled_sessions_count"`
	UnhandledRate          float64 `json:"unhandled_rate"`
	UsersSeen              int     `json:"users_seen"`
	UsersWithUnhandled     int     `json:"users_with_unhandled"`
	UnhandledUserRate      float64 `json:"unhandled_user_rate"`
}

func (t TimelinePoint) TableHeaders() []string {
	return []string{"BUCKET_START", "BUCKET_END", "SESSIONS", "UNHANDLED", "UNHANDLED_RATE"}
}

func (t TimelinePoint) TableRow() []string {
	return []string{t.BucketStart, t.BucketEnd, itoa(t.TotalSessionsCount), itoa(t.UnhandledSessionsCount), ftoa(t.UnhandledRate)}
}
