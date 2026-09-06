package parentchildpg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query     databasepg.Query `json:"query,omitempty" validate:"required"`
	Relation  *RelationType    `json:"relation,omitempty"`
	CanPickup *bool            `json:"can_pickup,omitempty"`
}

func (a UpdateArgument) Populate(parentChild ParentChild) ParentChild {

	if a.Relation != nil {
		parentChild.Relation = *a.Relation
	}

	if a.CanPickup != nil {
		parentChild.CanPickup = *a.CanPickup
	}

	return parentChild
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
