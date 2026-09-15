package authorizationpg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// RegisterParentArgument is a parent signing up entirely on their own, no
// director invite involved at all -- mobile-only client-side (per product
// decision: registration everywhere else stays as it was). Creates an
// Account with no OrganizationId, meant to be used for "personal" children
// (see project/service/internal/children/mine.go) and AI features; the
// parent can optionally link to a placowka later through the ordinary
// registration-link flow, same as any other parent account.
//
// Identity is proven exactly one of two ways -- Password OR a verified
// provider email -- same either-or rule and same reasoning as
// project/common/registrationlinks.PublicRegisterArgument.
//
// Name is deliberately optional here: the mobile onboarding flow collects
// email+password (or a provider token) first and asks for the parent's
// name on a separate follow-up screen, via PATCH /private/accounts, once
// the account already exists and the caller is authenticated.
type RegisterParentArgument struct {
	Name     string `json:"name,omitempty" validate:"omitempty,max=128"`
	Email    string `json:"email,omitempty" validate:"omitempty,email,max=128"`
	Password string `json:"password,omitempty" validate:"omitempty,min=6,max=72"`
	// VerifiedProviderEmail is set by the transport handler after
	// successfully verifying a Google/Apple token -- never trust a raw
	// client-submitted "I signed in with Google as X" claim.
	VerifiedProviderEmail string `json:"-"`
}

func (a RegisterParentArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	if a.Password == "" && a.VerifiedProviderEmail == "" {
		return errors.New("either password or a provider sign-in is required")
	}
	if a.Password != "" && a.VerifiedProviderEmail != "" {
		return errors.New("password and provider sign-in are mutually exclusive")
	}
	if a.Password != "" && a.Email == "" {
		return errors.New("Field 'Email' does not match validation: required\n")
	}

	return nil
}

func (a RegisterParentArgument) EffectiveEmail() string {
	if a.VerifiedProviderEmail != "" {
		return a.VerifiedProviderEmail
	}
	return a.Email
}

type RegisterParentRequest struct {
	Name               string `json:"name,omitempty"`
	Email              string `json:"email,omitempty"`
	Password           string `json:"password,omitempty"`
	GoogleIdToken      string `json:"google_id_token,omitempty"`
	AppleIdentityToken string `json:"apple_identity_token,omitempty"`
}
