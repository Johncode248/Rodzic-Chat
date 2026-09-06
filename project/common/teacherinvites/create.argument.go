package teacherinvitespg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const inviteValidity = 30 * 24 * time.Hour

type CreateArgument struct {
	Id             string `json:"id,omitempty"`
	OrganizationId string `json:"organization_id,omitempty" validate:"required"`
	GroupId        string `json:"group_id,omitempty"`
	// AccountType is "teacher" (default when empty) or "co_director" --
	// enforced by Validate() below.
	AccountType string `json:"account_type,omitempty" validate:"omitempty,oneof=teacher co_director"`
	Name        string `json:"name,omitempty" validate:"omitempty,max=128"`
	Email       string `json:"email,omitempty" validate:"omitempty,email,max=128"`
	CreatedBy   string `json:"created_by,omitempty" validate:"required"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (invite TeacherInvite) {

	invite.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	invite.OrganizationId = a.OrganizationId
	invite.GroupId = a.GroupId
	invite.Token = random.GenerateToken()
	invite.AccountType = expressions.Select(a.AccountType != "", a.AccountType, "teacher")
	invite.Name = a.Name
	invite.Email = a.Email
	invite.CreatedBy = a.CreatedBy
	invite.ExpiresAt = time.Now().Add(inviteValidity)
	invite.CreateTime = time.Now()
	invite.UpdateTime = time.Now()

	return invite
}
