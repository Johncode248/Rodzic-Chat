package teacherinvitespg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ConfirmRequest is POST /private/teacher-invites/confirm's body --
// authenticated, rate-limited. Attaches the CALLING account (from auth
// context) as staff of the invite's organization -- there is no more
// "create an account" branch, unlike the old public register flow this
// replaces.
type ConfirmRequest struct {
	Pin string `json:"pin,omitempty"`
}

type ConfirmArgument struct {
	Pin       string `json:"-" validate:"required"`
	AccountId string `json:"-" validate:"required"`
}

func (a ConfirmArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

type ConfirmResponse struct {
	Message string `json:"message,omitempty"`
}
