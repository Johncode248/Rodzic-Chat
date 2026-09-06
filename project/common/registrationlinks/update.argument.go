package registrationlinkspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

// UpdateArgument is used internally only (by the public registration
// handler, to atomically bump UsedCount) — there is no HTTP PATCH route
// for registration links.
type UpdateArgument struct {
	Query     databasepg.Query `json:"query,omitempty" validate:"required"`
	UsedCount *int             `json:"used_count,omitempty"`
}

func (a UpdateArgument) Populate(link RegistrationLink) RegistrationLink {

	link.UpdateTime = time.Now()

	if a.UsedCount != nil {
		link.UsedCount = *a.UsedCount
	}

	return link
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
