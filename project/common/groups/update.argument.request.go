package groupspg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type UpdateRequest struct {
	Query                             databasepg.Query `json:"query,omitempty"`
	Name                              *string          `json:"name,omitempty"`
	TeacherId                         *string          `json:"teacher_id,omitempty"`
	RequireChildBirthDateVerification *bool            `json:"require_child_birth_date_verification,omitempty"`
}
