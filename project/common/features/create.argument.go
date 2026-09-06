package featurespg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id             string  `json:"id,omitempty"`
	OrganizationId string  `json:"organization_id,omitempty" validate:"required"`
	Feature        Feature `json:"feature,omitempty" validate:"required"`
	Enabled        bool    `json:"enabled"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (organizationFeature OrganizationFeature) {

	organizationFeature.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	organizationFeature.OrganizationId = a.OrganizationId
	organizationFeature.Feature = a.Feature
	organizationFeature.Enabled = a.Enabled
	organizationFeature.CreateTime = time.Now()
	organizationFeature.UpdateTime = time.Now()

	return organizationFeature
}
