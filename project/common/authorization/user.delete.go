package authorizationpg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type DeleteUserArgument struct {
	UserId string `validate:"required,uuid"`
}

func (a DeleteUserArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
