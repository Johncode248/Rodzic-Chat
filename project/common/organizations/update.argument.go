package organizationspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query   databasepg.Query `json:"query,omitempty" validate:"required"`
	Name    *string          `json:"name,omitempty" validate:"omitempty,max=128"`
	Address *string          `json:"address,omitempty"`
	Nip     *string          `json:"nip,omitempty"`
	LogoUrl *string          `json:"logo_url,omitempty"`
	// ChildrenLimit must be one of ChildrenLimitOptions() (see pricing.go)
	// -- keep the "oneof" list below in sync with that slice. A drop below
	// the organization's current child count is additionally rejected in
	// project/service/internal/organizations.ValidateUpdateMutation (needs
	// a DB read, can't be checked here).
	ChildrenLimit *int    `json:"children_limit,omitempty" validate:"omitempty,oneof=20 30 40 50 80 100 120 200 300 400 500"`
	PrimaryColor  *string `json:"primary_color,omitempty"`
}

func (a UpdateArgument) Populate(organization Organization) Organization {

	organization.UpdateTime = time.Now()

	if a.Name != nil {
		organization.Name = *a.Name
	}

	if a.Address != nil {
		organization.Address = *a.Address
	}

	if a.Nip != nil {
		organization.Nip = *a.Nip
	}

	if a.LogoUrl != nil {
		organization.LogoUrl = *a.LogoUrl
	}

	if a.ChildrenLimit != nil {
		organization.ChildrenLimit = *a.ChildrenLimit
	}

	if a.PrimaryColor != nil {
		organization.PrimaryColor = *a.PrimaryColor
	}

	return organization
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
