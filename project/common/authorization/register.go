package authorizationpg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// RegisterArgument is the self-service "found a new placówka" flow — it is
// director-only. Parent and teacher accounts self-register through
// director-generated, expiring invite links instead — see
// project/common/registrationlinks (parents) and
// project/common/teacherinvites (teachers).
type RegisterArgument struct {
	Name                  string                 `json:"name,omitempty" validate:"required,max=128"`
	Email                 string                 `json:"email,omitempty" validate:"required,email,max=128"`
	Phone                 string                 `json:"phone,omitempty" validate:"required,max=32"`
	Password              string                 `json:"password,omitempty" validate:"required,max=72"`
	OrganizationName      string                 `json:"organization_name,omitempty" validate:"required,max=128"`
	OrganizationType      string                 `json:"organization_type,omitempty" validate:"required"`
	OrganizationSubdomain string                 `json:"organization_subdomain,omitempty" validate:"required,max=63,alphanum"`
	Data                  map[string]interface{} `json:"data,omitempty"`
}

func (a RegisterArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
