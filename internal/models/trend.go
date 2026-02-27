package models

type TrendBucket struct {
	From        string `json:"from"`
	To          string `json:"to"`
	EventsCount int    `json:"events_count"`
}

func (t TrendBucket) TableHeaders() []string {
	return []string{"FROM", "TO", "EVENTS_COUNT"}
}

func (t TrendBucket) TableRow() []string {
	return []string{t.From, t.To, itoa(t.EventsCount)}
}
