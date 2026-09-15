package childrenpg

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// CreateMineRequest/Argument: a parent adding a child entirely on their own,
// independent of any placowka -- see project/service/internal/children/mine.go.
// Unlike the staff-facing CreateRequest (children.Create), there is
// deliberately no OrganizationId/GroupId here: this child starts out
// belonging to no organization at all (personal use -- AI features), and
// can optionally be linked to a placowka later through the ordinary
// registration-link flow, exactly like any other child.
type CreateMineRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	// BirthDate as "2006-01-02".
	BirthDate string `json:"birth_date,omitempty"`
}

type CreateMineArgument struct {
	ParentAccountId string    `json:"-" validate:"required"`
	FirstName       string    `json:"first_name,omitempty" validate:"required,max=64"`
	LastName        string    `json:"last_name,omitempty" validate:"required,max=64"`
	BirthDate       time.Time `json:"birth_date,omitempty"`
}

func (a CreateMineArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type CreateMineResponse struct {
	Child Child `json:"child,omitempty"`
}

// SelectMineResponse is every child a parent can see, whether personal
// (OrganizationId == "") or connected to a placowka -- one unified list,
// see project/service/internal/children/mine.go.
type SelectMineResponse struct {
	Children []Child `json:"children,omitempty"`
}
