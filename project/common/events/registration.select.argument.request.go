package eventspg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type RegistrationSelectRequest struct {
	Query databasepg.Query `json:"query,omitempty"`
}
