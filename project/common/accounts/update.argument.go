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
	// Permissions is set only by the director-only staff-permissions
	// route (accountsgt.UpdateStaffPermissions) -- it is deliberately
	// absent from UpdateRequest, the DTO the self-service PATCH
	// /private/accounts endpoint decodes into, so a caller can never grant
	// themselves permissions through the self-update path.
	Permissions *Permission `json:"permissions,omitempty"`
}

func (a UpdateArgument) Populate(account Account) Account {

	account.UpdateTime = time.Now()

	if a.Name != nil {
		account.Name = *a.Name
	}

	if a.Phone != nil {
		account.Phone = *a.Phone
	}

	if a.Permissions != nil {
		account.Permissions = *a.Permissions
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
