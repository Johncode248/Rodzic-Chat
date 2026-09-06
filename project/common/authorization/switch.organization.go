package authorizationpg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type SwitchOrganizationArgument struct {
	AccountId      string `json:"-" validate:"required"`
	Email          string `json:"-" validate:"required"`
	SessionId      string `json:"-" validate:"required"`
	OrganizationId string `json:"organization_id,omitempty" validate:"required"`
}

func (a SwitchOrganizationArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
