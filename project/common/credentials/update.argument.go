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
// "omitempty" here only skips validation when Password itself is nil (the
// caller isn't changing their password at all) -- a non-nil pointer, even
// to "", is still checked against min=6, so a caller can no longer set
// their password to an empty string by sending {"password":""}.
type UpdateArgument struct {
	Query databasepg.Query `json:"query,omitempty"`
	// UserId is who the update is for -- used only as a fallback to CREATE a
	// Credential row when none exists yet (see credentials.Update), since
	// Query is a generic filter set the create path can't derive an id from.
	UserId   string  `json:"user_id,omitempty"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=6,max=72"`
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
