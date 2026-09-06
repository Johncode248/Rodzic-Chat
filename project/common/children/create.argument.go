package childrenpg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id             string    `json:"id,omitempty"`
	OrganizationId string    `json:"organization_id,omitempty" validate:"required"`
	GroupId        string    `json:"group_id,omitempty"`
	FirstName      string    `json:"first_name,omitempty" validate:"required,max=128"`
	LastName       string    `json:"last_name,omitempty" validate:"required,max=128"`
	BirthDate      time.Time `json:"birth_date,omitempty"`
	Notes          string    `json:"notes,omitempty"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (child Child) {

	child.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	child.OrganizationId = a.OrganizationId
	child.GroupId = a.GroupId
	child.FirstName = a.FirstName
	child.LastName = a.LastName
	child.BirthDate = a.BirthDate
	child.Notes = a.Notes
	child.CreateTime = time.Now()

	return child
}
