package passwordresetspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query  databasepg.Query `json:"query,omitempty" validate:"required"`
	UsedAt *time.Time       `json:"used_at,omitempty"`
}

func (a UpdateArgument) Populate(reset PasswordReset) PasswordReset {

	if a.UsedAt != nil {
		reset.UsedAt = a.UsedAt
	}

	return reset
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
