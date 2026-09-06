package featurespg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query   databasepg.Query `json:"query,omitempty" validate:"required"`
	Enabled *bool            `json:"enabled,omitempty"`
}

func (a UpdateArgument) Populate(organizationFeature OrganizationFeature) OrganizationFeature {

	organizationFeature.UpdateTime = time.Now()

	if a.Enabled != nil {
		organizationFeature.Enabled = *a.Enabled
	}

	return organizationFeature
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
