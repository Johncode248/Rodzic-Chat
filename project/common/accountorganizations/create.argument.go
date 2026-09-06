package accountorganizationspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id             string `json:"id,omitempty"`
	AccountId      string `json:"account_id,omitempty" validate:"required"`
	OrganizationId string `json:"organization_id,omitempty" validate:"required"`
	IsPrimary      bool   `json:"is_primary,omitempty"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (membership AccountOrganization) {

	membership.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	membership.AccountId = a.AccountId
	membership.OrganizationId = a.OrganizationId
	membership.IsPrimary = a.IsPrimary
	membership.CreateTime = time.Now()

	return membership
}
