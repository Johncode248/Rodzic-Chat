package accountspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
	Name  *string          `json:"name,omitempty" validate:"omitempty,max=64"`
	Phone *string          `json:"phone,omitempty" validate:"omitempty,max=20"`
}

func (a UpdateArgument) Populate(account Account) Account {

	account.UpdateTime = time.Now()

	if a.Name != nil {
		account.Name = *a.Name
	}

	if a.Phone != nil {
		account.Phone = *a.Phone
	}

	return account
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
