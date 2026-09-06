package credentialspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// Password is capped at 72 bytes -- bcrypt.GenerateFromPassword errors on
// anything longer, and Populate() below silently discards that error by
// returning a zero-value Credential (a pre-existing footgun in this file);
// capping the input here is what actually prevents that from firing.
type CreateArgument struct {
	UserId   string `json:"user_id,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required,max=72"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (credential Credential) {

	credential.CreateTime = time.Now()

	if a.UserId != "" {
		credential.UserId = a.UserId
	}

	if a.Password != "" {
		password, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
		if err != nil {
			return Credential{}
		}

		credential.Password = string(password)
	}

	return credential
}
