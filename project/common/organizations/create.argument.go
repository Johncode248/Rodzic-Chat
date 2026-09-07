package organizationspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id        string           `json:"id,omitempty"`
	Name      string           `json:"name,omitempty" validate:"required,max=128"`
	Type      OrganizationType `json:"type,omitempty" validate:"required"`
	Subdomain string           `json:"subdomain,omitempty" validate:"required,max=63,alphanum"`
	Address   string           `json:"address,omitempty"`
	Nip       string           `json:"nip,omitempty"`
	LogoUrl   string           `json:"logo_url,omitempty"`
	// ChildrenLimit must be one of ChildrenLimitOptions() (see pricing.go)
	// -- keep the "oneof" list below in sync with that slice.
	ChildrenLimit int    `json:"children_limit,omitempty" validate:"required,oneof=20 30 40 50 80 100 120 200 300 400 500"`
	PrimaryColor  string `json:"primary_color,omitempty"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (organization Organization) {

	organization.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	organization.Name = a.Name
	organization.Type = a.Type
	organization.Subdomain = a.Subdomain
	organization.Address = a.Address
	organization.Nip = a.Nip
	organization.LogoUrl = a.LogoUrl
	organization.PrimaryColor = a.PrimaryColor
	organization.ChildrenLimit = a.ChildrenLimit
	organization.PaymentReference = random.GeneratePaymentReference()
	organization.SubscriptionStatus = SubscriptionStatusTrial
	organization.CreateTime = time.Now()

	return organization
}
