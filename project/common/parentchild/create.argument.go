package parentchildpg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id        string       `json:"id,omitempty"`
	ParentId  string       `json:"parent_id,omitempty" validate:"required"`
	ChildId   string       `json:"child_id,omitempty" validate:"required"`
	Relation  RelationType `json:"relation,omitempty"`
	CanPickup bool         `json:"can_pickup,omitempty"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (parentChild ParentChild) {

	parentChild.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	parentChild.ParentId = a.ParentId
	parentChild.ChildId = a.ChildId
	parentChild.Relation = a.Relation
	parentChild.CanPickup = a.CanPickup
	parentChild.CreateTime = time.Now()

	return parentChild
}
