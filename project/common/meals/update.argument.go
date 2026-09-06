package mealspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query    databasepg.Query `json:"query,omitempty" validate:"required"`
	DateFrom *time.Time       `json:"date_from,omitempty"`
	DateTo   *time.Time       `json:"date_to,omitempty"`
	Reason   *string          `json:"reason,omitempty"`
}

func (a UpdateArgument) Populate(cancellation MealCancellation) MealCancellation {

	cancellation.UpdateTime = time.Now()

	if a.DateFrom != nil {
		cancellation.DateFrom = *a.DateFrom
	}

	if a.DateTo != nil {
		cancellation.DateTo = *a.DateTo
	}

	if a.Reason != nil {
		cancellation.Reason = *a.Reason
	}

	return cancellation
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
