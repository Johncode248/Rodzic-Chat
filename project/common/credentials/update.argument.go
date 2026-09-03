package credentialspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"golang.org/x/crypto/bcrypt"
)

type UpdateArgument struct {
	Query    databasepg.Query `json:"query,omitempty"`
	Password *string          `json:"password,omitempty"`
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
