package parentchildpg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type UpdateRequest struct {
	Query    databasepg.Query `json:"query,omitempty"`
	Relation *RelationType    `json:"relation,omitempty"`
}
