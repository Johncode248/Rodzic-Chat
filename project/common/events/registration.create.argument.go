package eventspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type RegistrationCreateArgument struct {
	Id          string             `json:"id,omitempty"`
	EventId     string             `json:"event_id,omitempty" validate:"required"`
	ChildId     string             `json:"child_id,omitempty" validate:"required"`
	ConfirmedBy string             `json:"confirmed_by,omitempty" validate:"required"`
	Status      RegistrationStatus `json:"status,omitempty" validate:"required"`
}

func (a RegistrationCreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a RegistrationCreateArgument) Populate() (registration EventRegistration) {

	registration.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	registration.EventId = a.EventId
	registration.ChildId = a.ChildId
	registration.ConfirmedBy = a.ConfirmedBy
	registration.Status = a.Status
	registration.CreateTime = time.Now()
	registration.UpdateTime = time.Now()

	return registration
}
