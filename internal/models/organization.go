package models

type Organization struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	AutoUpgrade   bool   `json:"auto_upgrade"`
	BillingEmails []string `json:"billing_emails,omitempty"`
}

func (o Organization) TableHeaders() []string {
	return []string{"ID", "NAME", "SLUG", "CREATED_AT"}
}

func (o Organization) TableRow() []string {
	return []string{o.ID, o.Name, o.Slug, o.CreatedAt}
}
