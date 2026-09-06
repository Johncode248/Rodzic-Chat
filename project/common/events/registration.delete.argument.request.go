package eventspg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type RegistrationDeleteRequest struct {
	Query databasepg.Query `json:"query,omitempty"`
}
