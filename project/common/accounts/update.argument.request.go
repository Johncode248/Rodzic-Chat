package accountspg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type UpdateRequest struct {
	Query databasepg.Query `json:"query,omitempty"`
	Name  *string          `json:"name,omitempty"`
	Phone *string          `json:"phone,omitempty"`
}
