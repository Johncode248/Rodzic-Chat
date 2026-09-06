package childrenpg

import (
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
)

type UpdateRequest struct {
	Query     databasepg.Query `json:"query,omitempty"`
	GroupId   *string          `json:"group_id,omitempty"`
	FirstName *string          `json:"first_name,omitempty"`
	LastName  *string          `json:"last_name,omitempty"`
	BirthDate *time.Time       `json:"birth_date,omitempty"`
	Notes     *string          `json:"notes,omitempty"`
}
