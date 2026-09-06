package credentialspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// Password is capped at 72 bytes -- see the identical comment in
// create.argument.go; Populate() below has the same silent-zero-value
// footgun on a bcrypt error, so validation here is the real guard.
type UpdateArgument struct {
	Query    databasepg.Query `json:"query,omitempty"`
	Password *string          `json:"password,omitempty" validate:"omitempty,max=72"`
}

func (a UpdateArgument) Populate(credential Credential) Credential {

	credential.UpdateTime = time.Now()

	if a.Password != nil {
		password, err := bcrypt.GenerateFromPassword([]byte(*a.Password), bcrypt.DefaultCost)
		if err != nil {
			return Credential{}
		}

		credential.Password = string(password)
	}

	return credential
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
