package groupspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query                             databasepg.Query `json:"query,omitempty" validate:"required"`
	Name                              *string          `json:"name,omitempty" validate:"omitempty,max=128"`
	TeacherId                         *string          `json:"teacher_id,omitempty"`
	RequireChildBirthDateVerification *bool            `json:"require_child_birth_date_verification,omitempty"`
}

func (a UpdateArgument) Populate(group Group) Group {

	group.UpdateTime = time.Now()

	if a.Name != nil {
		group.Name = *a.Name
	}

	if a.TeacherId != nil {
		group.TeacherId = *a.TeacherId
	}

	if a.RequireChildBirthDateVerification != nil {
		group.RequireChildBirthDateVerification = *a.RequireChildBirthDateVerification
	}

	return group
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
