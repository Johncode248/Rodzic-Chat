package registrationlinkspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const linkValidity = 72 * time.Hour

type CreateArgument struct {
	Id                string `json:"id,omitempty"`
	OrganizationId    string `json:"organization_id,omitempty" validate:"required"`
	GroupId           string `json:"group_id,omitempty" validate:"required"`
	ChildrenCount     int    `json:"children_count,omitempty" validate:"required,min=1,max=200"`
	PublicDescription string `json:"public_description,omitempty" validate:"omitempty,max=256"`
	InternalNote      string `json:"internal_note,omitempty" validate:"omitempty,max=1024"`
	CreatedBy         string `json:"created_by,omitempty" validate:"required"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (link RegistrationLink) {

	link.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	link.OrganizationId = a.OrganizationId
	link.GroupId = a.GroupId
	link.ChildrenCount = a.ChildrenCount
	link.MaxRegistrations = a.ChildrenCount * 2
	link.UsedCount = 0
	link.PublicDescription = a.PublicDescription
	link.InternalNote = a.InternalNote
	link.CreatedBy = a.CreatedBy
	link.ExpiresAt = time.Now().Add(linkValidity)
	link.CreateTime = time.Now()
	link.UpdateTime = time.Now()

	// PinHash is set by the caller (see registrationlinks.Create) after
	// generating and hashing a fresh plaintext PIN -- Populate has no way
	// to hand the plaintext value back to its caller, and the plaintext
	// must never be stored, only returned once in the create response.

	return link
}
