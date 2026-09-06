package teacherinvitespg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

// UpdateArgument is used internally only (by the public register handler,
// to mark an invite used) — there is no HTTP PATCH route for teacher
// invites.
type UpdateArgument struct {
	Query  databasepg.Query `json:"query,omitempty" validate:"required"`
	UsedAt *time.Time       `json:"used_at,omitempty"`
}

func (a UpdateArgument) Populate(invite TeacherInvite) TeacherInvite {

	invite.UpdateTime = time.Now()

	if a.UsedAt != nil {
		invite.UsedAt = a.UsedAt
	}

	return invite
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
