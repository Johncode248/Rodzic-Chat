package childrenpg

import "time"

type CreateRequest struct {
	GroupId   string    `json:"group_id,omitempty"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	BirthDate time.Time `json:"birth_date,omitempty"`
	Notes     string    `json:"notes,omitempty"`
}
