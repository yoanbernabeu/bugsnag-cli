package models

type Comment struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	AuthorID  string `json:"author_id"`
	AuthorName string `json:"author_name"`
	CreatedAt string `json:"created_at"`
}

func (c Comment) TableHeaders() []string {
	return []string{"ID", "AUTHOR", "MESSAGE", "CREATED_AT"}
}

func (c Comment) TableRow() []string {
	msg := c.Message
	if len(msg) > 60 {
		msg = msg[:57] + "..."
	}
	return []string{c.ID, c.AuthorName, msg, c.CreatedAt}
}
