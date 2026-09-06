package parentchildpg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type DeleteRequest struct {
	Query databasepg.Query `json:"query,omitempty"`
}
