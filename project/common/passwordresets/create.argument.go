package passwordresetspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id        string    `json:"id,omitempty"`
	AccountId string    `json:"account_id,omitempty" validate:"required"`
	TokenHash string    `json:"-" validate:"required"`
	ExpiresAt time.Time `json:"expires_at,omitempty" validate:"required"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (reset PasswordReset) {

	reset.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	reset.AccountId = a.AccountId
	reset.TokenHash = a.TokenHash
	reset.ExpiresAt = a.ExpiresAt
	reset.CreateTime = time.Now()

	return reset
}
