package teacherinvitespg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type PublicRegisterRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// PublicRegisterArgument is how a teacher completes their account through a
// director-generated invite link. Token is never trusted as an
// authorization mechanism beyond "which invite" — expiry/single-use and
// email-not-already-taken are enforced server-side in
// project/service/internal/teacherinvites.
type PublicRegisterArgument struct {
	Token    string `json:"-" validate:"required"`
	Name     string `json:"name,omitempty" validate:"required,max=128"`
	Email    string `json:"email,omitempty" validate:"required,email,max=128"`
	Password string `json:"password,omitempty" validate:"required,min=6,max=72"`
}

func (a PublicRegisterArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

type PublicRegisterResponse struct {
	Message string `json:"message,omitempty"`
}
