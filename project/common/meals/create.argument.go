package mealspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id             string    `json:"id,omitempty"`
	OrganizationId string    `json:"organization_id,omitempty" validate:"required"`
	ChildId        string    `json:"child_id,omitempty" validate:"required"`
	ReportedBy     string    `json:"reported_by,omitempty" validate:"required"`
	DateFrom       time.Time `json:"date_from,omitempty" validate:"required"`
	DateTo         time.Time `json:"date_to,omitempty" validate:"required"`
	Reason         string    `json:"reason,omitempty"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	if a.DateTo.Before(a.DateFrom) {
		return errors.New("date_to must not be before date_from")
	}

	return nil
}

func (a CreateArgument) Populate() (cancellation MealCancellation) {

	cancellation.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	cancellation.OrganizationId = a.OrganizationId
	cancellation.ChildId = a.ChildId
	cancellation.ReportedBy = a.ReportedBy
	cancellation.DateFrom = a.DateFrom
	cancellation.DateTo = a.DateTo
	cancellation.Reason = a.Reason
	cancellation.CreateTime = time.Now()
	cancellation.UpdateTime = time.Now()

	return cancellation
}
