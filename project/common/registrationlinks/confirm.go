package registrationlinkspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// ConfirmRequest is POST /private/registration-invites/confirm's body --
// authenticated, rate-limited. Attaches the CALLING account (from auth
// context, not a newly-created one) as a parent, since the person is
// already logged in -- there is no more "create an account" branch, unlike
// the old public register flow this replaces.
type ConfirmRequest struct {
	Pin string `json:"pin,omitempty"`
	// ChildId picks one of LookupResponse.Children instead of typing a
	// name below -- mutually exclusive with ChildFirstName/ChildLastName
	// (see ConfirmArgument.Validate).
	ChildId        string `json:"child_id,omitempty"`
	ChildFirstName string `json:"child_first_name,omitempty"`
	ChildLastName  string `json:"child_last_name,omitempty"`
	// ChildBirthDate as "2006-01-02" -- required only when the target
	// group has RequireChildBirthDateVerification enabled (checked
	// server-side, not by this struct's shape, since it depends on which
	// group the invite belongs to).
	ChildBirthDate string `json:"child_birth_date,omitempty"`
}

// ConfirmArgument is how an already-authenticated parent completes
// attaching to a child through a director-generated PIN. The PIN is never
// trusted as an authorization mechanism beyond "which invite" — every
// other constraint (expiry, remaining capacity, child already has two
// parents) is enforced server-side in project/service/internal/registrationlinks.
type ConfirmArgument struct {
	Pin       string `json:"-" validate:"required"`
	AccountId string `json:"-" validate:"required"`
	// ChildId XOR (ChildFirstName+ChildLastName) -- see Validate.
	ChildId        string     `json:"child_id,omitempty"`
	ChildFirstName string     `json:"child_first_name,omitempty" validate:"omitempty,max=64"`
	ChildLastName  string     `json:"child_last_name,omitempty" validate:"omitempty,max=64"`
	ChildBirthDate *time.Time `json:"child_birth_date,omitempty"`
}

func (a ConfirmArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	// Exactly one of "pick an existing child" or "type a new child's name"
	// -- see ChildId's doc comment.
	if a.ChildId == "" && (a.ChildFirstName == "" || a.ChildLastName == "") {
		return errors.New("either child_id or both child_first_name and child_last_name are required")
	}
	if a.ChildId != "" && (a.ChildFirstName != "" || a.ChildLastName != "") {
		return errors.New("child_id and child_first_name/child_last_name are mutually exclusive")
	}

	return nil
}

type ConfirmResponse struct {
	Message string `json:"message,omitempty"`
	// PendingApproval is true when confirming created a GuardianClaim
	// instead of an immediate ParentChild link -- the caller's account
	// already exists and is already logged in, but has no access to the
	// child yet (see project/common/guardianclaims).
	PendingApproval bool `json:"pending_approval,omitempty"`
}
