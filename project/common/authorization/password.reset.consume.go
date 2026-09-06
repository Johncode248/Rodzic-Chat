package authorizationpg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ConsumePasswordResetArgument struct {
	Token       string `json:"token,omitempty" validate:"required"`
	NewPassword string `json:"new_password,omitempty" validate:"required,min=6,max=72"`
}

func (a ConsumePasswordResetArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
