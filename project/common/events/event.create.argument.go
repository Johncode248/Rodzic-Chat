package eventspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type EventCreateArgument struct {
	Id                   string     `json:"id,omitempty"`
	OrganizationId       string     `json:"organization_id,omitempty" validate:"required"`
	GroupId              string     `json:"group_id,omitempty"`
	Type                 EventType  `json:"type,omitempty" validate:"required"`
	Title                string     `json:"title,omitempty" validate:"required,max=128"`
	Description          string     `json:"description,omitempty"`
	RequiresItem         string     `json:"requires_item,omitempty"`
	EventDate            time.Time  `json:"event_date,omitempty" validate:"required"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	CreatedBy            string     `json:"created_by,omitempty" validate:"required"`
}

func (a EventCreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a EventCreateArgument) Populate() (event Event) {

	event.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	event.OrganizationId = a.OrganizationId
	event.GroupId = a.GroupId
	event.Type = a.Type
	event.Title = a.Title
	event.Description = a.Description
	event.RequiresItem = a.RequiresItem
	event.EventDate = a.EventDate
	event.RegistrationDeadline = a.RegistrationDeadline
	event.CreatedBy = a.CreatedBy
	event.CreateTime = time.Now()
	event.UpdateTime = time.Now()

	return event
}
