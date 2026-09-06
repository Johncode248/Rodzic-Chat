package featurespg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type SelectRequest struct {
	Query databasepg.Query `json:"query,omitempty"`
}
