package eventspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type EventUpdateArgument struct {
	Query                databasepg.Query `json:"query,omitempty" validate:"required"`
	Title                *string          `json:"title,omitempty" validate:"omitempty,max=128"`
	Description          *string          `json:"description,omitempty"`
	RequiresItem         *string          `json:"requires_item,omitempty"`
	EventDate            *time.Time       `json:"event_date,omitempty"`
	RegistrationDeadline *time.Time       `json:"registration_deadline,omitempty"`
}

func (a EventUpdateArgument) Populate(event Event) Event {

	event.UpdateTime = time.Now()

	if a.Title != nil {
		event.Title = *a.Title
	}

	if a.Description != nil {
		event.Description = *a.Description
	}

	if a.RequiresItem != nil {
		event.RequiresItem = *a.RequiresItem
	}

	if a.EventDate != nil {
		event.EventDate = *a.EventDate
	}

	if a.RegistrationDeadline != nil {
		event.RegistrationDeadline = a.RegistrationDeadline
	}

	return event
}

func (a EventUpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
