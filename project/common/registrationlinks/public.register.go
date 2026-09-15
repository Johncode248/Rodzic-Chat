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
	// Email is required when none of the provider tokens below are set --
	// when a provider token IS set, whatever verified email it carries
	// wins, and this field (if also sent) is ignored rather than trusted.
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	// GoogleIdToken/AppleIdentityToken: exactly one of {Password,
	// GoogleIdToken, AppleIdentityToken} must be set -- see
	// PublicRegisterArgument's doc comment.
	GoogleIdToken      string `json:"google_id_token,omitempty"`
	AppleIdentityToken string `json:"apple_identity_token,omitempty"`
}

// PublicRegisterArgument is how a parent completes self-registration
// through a director-generated link. Token is never trusted as an
// authorization mechanism beyond "which link" — every other constraint
// (expiry, remaining capacity, no duplicate name within the group, email
// not already in use) is enforced server-side in
// project/service/internal/registrationlinks.
//
// Identity is proven exactly one of three ways -- Password OR GoogleIdToken
// OR AppleIdentityToken, never a mix, enforced in Validate() below since
// that's an either-or rule struct tags alone can't express:
//   - Password: verified against bcrypt as before (email typed by hand).
//   - GoogleIdToken/AppleIdentityToken: verified server-side against the
//     provider's own signing keys (project/service/internal/oauthverify)
//     BEFORE this argument is even built (see the transport handler) --
//     the resulting verified email overwrites whatever Email was submitted,
//     so Email becomes optional/ignored in that case.
//
// Either way, when the submitted/verified email matches an existing
// account, ownership is still proven (bcrypt password check, or the
// provider's own verified-email guarantee) rather than skipped — so
// joining a second facility can't be done by anyone who merely knows the
// account's email address (see PublicRegister's joiningExistingAccount
// branch). A parent who forgot their password uses the ordinary
// /public/password-recovery flow first (or just signs in with a provider,
// if their account's email matches one).
type PublicRegisterArgument struct {
	Token          string     `json:"-" validate:"required"`
	ChildFirstName string     `json:"child_first_name,omitempty" validate:"required,max=64"`
	ChildLastName  string     `json:"child_last_name,omitempty" validate:"required,max=64"`
	ChildBirthDate *time.Time `json:"child_birth_date,omitempty"`
	Name           string     `json:"name,omitempty" validate:"required,max=128"`
	Email          string     `json:"email,omitempty" validate:"omitempty,email,max=128"`
	Password       string     `json:"password,omitempty" validate:"omitempty,min=6,max=72"`
	// VerifiedProviderEmail is set by the transport handler after
	// successfully verifying a Google/Apple token -- never trust a raw
	// client-submitted "I signed in with Google as X" claim. Internal-only
	// (json:"-"), mutually exclusive with Password.
	VerifiedProviderEmail string `json:"-"`
}

func (a PublicRegisterArgument) Validate() error {

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

// EffectiveEmail is VerifiedProviderEmail when set (provider sign-in),
// otherwise the hand-typed Email (password sign-in) -- always use this,
// never Email directly, once past Validate().
func (a PublicRegisterArgument) EffectiveEmail() string {
	if a.VerifiedProviderEmail != "" {
		return a.VerifiedProviderEmail
	}
	return a.Email
}

type PublicRegisterResponse struct {
	Message string `json:"message,omitempty"`
	// PendingApproval is true when the registration created a
	// GuardianClaim instead of an immediate ParentChild link -- the
	// account exists and can log in, but has no access to the child yet
	// (see project/common/guardianclaims).
	PendingApproval bool `json:"pending_approval,omitempty"`
}
