package eventspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type RegistrationUpdateArgument struct {
	Query       databasepg.Query    `json:"query,omitempty" validate:"required"`
	ConfirmedBy string              `json:"confirmed_by,omitempty" validate:"required"`
	Status      *RegistrationStatus `json:"status,omitempty"`
}

func (a RegistrationUpdateArgument) Populate(registration EventRegistration) EventRegistration {

	registration.UpdateTime = time.Now()
	registration.ConfirmedBy = a.ConfirmedBy

	if a.Status != nil {
		registration.Status = *a.Status
	}

	return registration
}

func (a RegistrationUpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
