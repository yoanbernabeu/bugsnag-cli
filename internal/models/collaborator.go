package models

type Collaborator struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsAdmin   bool   `json:"is_admin"`
	ProjectsCount int `json:"projects_count"`
	CreatedAt string `json:"created_at"`
}

func (c Collaborator) TableHeaders() []string {
	return []string{"ID", "NAME", "EMAIL", "IS_ADMIN"}
}

func (c Collaborator) TableRow() []string {
	admin := "no"
	if c.IsAdmin {
		admin = "yes"
	}
	return []string{c.ID, c.Name, c.Email, admin}
}
