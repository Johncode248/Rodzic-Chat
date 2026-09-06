package absencespg

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

func (a UpdateArgument) Populate(absence Absence) Absence {

	absence.UpdateTime = time.Now()

	if a.DateFrom != nil {
		absence.DateFrom = *a.DateFrom
	}

	if a.DateTo != nil {
		absence.DateTo = *a.DateTo
	}

	if a.Reason != nil {
		absence.Reason = *a.Reason
	}

	return absence
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
