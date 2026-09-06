package childrenpg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query     databasepg.Query `json:"query,omitempty" validate:"required"`
	GroupId   *string          `json:"group_id,omitempty"`
	FirstName *string          `json:"first_name,omitempty" validate:"omitempty,max=128"`
	LastName  *string          `json:"last_name,omitempty" validate:"omitempty,max=128"`
	BirthDate *time.Time       `json:"birth_date,omitempty"`
	Notes     *string          `json:"notes,omitempty"`
}

func (a UpdateArgument) Populate(child Child) Child {

	child.UpdateTime = time.Now()

	if a.GroupId != nil {
		child.GroupId = *a.GroupId
	}

	if a.FirstName != nil {
		child.FirstName = *a.FirstName
	}

	if a.LastName != nil {
		child.LastName = *a.LastName
	}

	if a.BirthDate != nil {
		child.BirthDate = *a.BirthDate
	}

	if a.Notes != nil {
		child.Notes = *a.Notes
	}

	return child
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
