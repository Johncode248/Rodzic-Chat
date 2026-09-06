package registrationlinkspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type PublicRegisterRequest struct {
	ChildFirstName string `json:"child_first_name,omitempty"`
	ChildLastName  string `json:"child_last_name,omitempty"`
	// ChildBirthDate as "2006-01-02" -- required only when the target
	// group has RequireChildBirthDateVerification enabled (checked
	// server-side, not by this struct's shape, since it depends on which
	// group the link belongs to).
	ChildBirthDate string `json:"child_birth_date,omitempty"`
	Name           string `json:"name,omitempty"`
	Email          string `json:"email,omitempty"`
	Password       string `json:"password,omitempty"`
}

// PublicRegisterArgument is how a parent completes self-registration
// through a director-generated link. Token is never trusted as an
// authorization mechanism beyond "which link" — every other constraint
// (expiry, remaining capacity, no duplicate name within the group, email
// not already in use) is enforced server-side in
// project/service/internal/registrationlinks. Password is ALWAYS required
// now, even when the submitted email matches an existing account: it's
// verified against that account's existing credential instead of being
// skipped, so joining a second facility can't be done by anyone who merely
// knows the account's email address (see PublicRegister's
// joiningExistingAccount branch) -- a parent who forgot their password
// uses the ordinary /public/password-recovery flow first.
type PublicRegisterArgument struct {
	Token          string     `json:"-" validate:"required"`
	ChildFirstName string     `json:"child_first_name,omitempty" validate:"required,max=64"`
	ChildLastName  string     `json:"child_last_name,omitempty" validate:"required,max=64"`
	ChildBirthDate *time.Time `json:"child_birth_date,omitempty"`
	Name           string     `json:"name,omitempty" validate:"required,max=128"`
	Email          string     `json:"email,omitempty" validate:"required,email,max=128"`
	Password       string     `json:"password,omitempty" validate:"required,min=6,max=72"`
}

func (a PublicRegisterArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

type PublicRegisterResponse struct {
	Message string `json:"message,omitempty"`
	// PendingApproval is true when the registration created a
	// GuardianClaim instead of an immediate ParentChild link -- the
	// account exists and can log in, but has no access to the child yet
	// (see project/common/guardianclaims).
	PendingApproval bool `json:"pending_approval,omitempty"`
}
