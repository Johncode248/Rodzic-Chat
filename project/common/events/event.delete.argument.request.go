package eventspg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type EventDeleteRequest struct {
	Query databasepg.Query `json:"query,omitempty"`
}
