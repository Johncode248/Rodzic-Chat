package authorizationpg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type RegisterArgument struct {
	Name     string                 `json:"name,omitempty" validate:"required,max=128"`
	Email    string                 `json:"email,omitempty" validate:"required,email,max=128"`
	Phone    string                 `json:"phone,omitempty" validate:"required,max=32"`
	Password string                 `json:"password,omitempty" validate:"required,max=128"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

func (a RegisterArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
