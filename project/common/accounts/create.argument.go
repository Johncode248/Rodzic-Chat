package accountspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id             string      `json:"id,omitempty"`
	OrganizationId string      `json:"organization_id,omitempty"`
	Name           string      `json:"name,omitempty" validate:"required,max=128"`
	Login          string      `json:"login,omitempty" validate:"required,max=128"`
	Email          string      `json:"email,omitempty" validate:"omitempty,email,max=128"`
	Phone          string      `json:"phone,omitempty"`
	AccountType    AccountType `json:"account_type,omitempty"`
	Status         Status      `json:"status,omitempty"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (account Account) {

	account.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	account.OrganizationId = a.OrganizationId
	account.Name = a.Name
	account.Login = a.Login
	account.Email = a.Email
	account.Phone = a.Phone
	account.AccountType = a.AccountType
	account.Status = expressions.Select(a.Status != StatusUnknown, a.Status, StatusEnabled)
	account.CreateTime = time.Now()

	return account
}
