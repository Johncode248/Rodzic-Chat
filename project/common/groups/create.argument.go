package groupspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id                                string `json:"id,omitempty"`
	OrganizationId                    string `json:"organization_id,omitempty" validate:"required"`
	Name                              string `json:"name,omitempty" validate:"required,max=128"`
	TeacherId                         string `json:"teacher_id,omitempty"`
	RequireChildBirthDateVerification bool   `json:"require_child_birth_date_verification,omitempty"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (group Group) {

	group.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	group.OrganizationId = a.OrganizationId
	group.Name = a.Name
	group.TeacherId = a.TeacherId
	group.RequireChildBirthDateVerification = a.RequireChildBirthDateVerification
	group.CreateTime = time.Now()

	return group
}
