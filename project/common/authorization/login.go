package authorizationpg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type LoginArgument struct {
	Email    string `validate:"required,email,max=128"`
	Password string `validate:"required,min=6"`
}

func (a LoginArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
