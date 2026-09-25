package authorizationpg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password,omitempty"`
}

type ChangePasswordResponse struct{}

// ChangePasswordArgument -- the new-password rule (6..72 bytes; 72 is bcrypt's hard limit) is the
// same one ConsumePasswordResetArgument enforces, so both paths accept exactly the same passwords.
type ChangePasswordArgument struct {
	AccountId       string `json:"account_id,omitempty" validate:"required"`
	SessionId       string `json:"session_id,omitempty"`
	CurrentPassword string `json:"current_password,omitempty" validate:"required"`
	NewPassword     string `json:"new_password,omitempty" validate:"required,min=6,max=72"`
}

func (a ChangePasswordArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}
