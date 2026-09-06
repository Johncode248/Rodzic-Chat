package guardianclaimspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id                   string    `json:"id,omitempty"`
	OrganizationId       string    `json:"organization_id,omitempty" validate:"required"`
	ChildId              string    `json:"child_id,omitempty" validate:"required"`
	ClaimantId           string    `json:"claimant_id,omitempty" validate:"required"`
	ClaimantName         string    `json:"claimant_name,omitempty" validate:"required"`
	ClaimantEmail        string    `json:"claimant_email,omitempty" validate:"required"`
	ClaimantBirthDate    time.Time `json:"claimant_birth_date,omitempty" validate:"required"`
	ChildBirthDateOnFile time.Time `json:"child_birth_date_on_file,omitempty" validate:"required"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (claim GuardianClaim) {

	claim.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	claim.OrganizationId = a.OrganizationId
	claim.ChildId = a.ChildId
	claim.ClaimantId = a.ClaimantId
	claim.ClaimantName = a.ClaimantName
	claim.ClaimantEmail = a.ClaimantEmail
	claim.ClaimantBirthDate = a.ClaimantBirthDate
	claim.ChildBirthDateOnFile = a.ChildBirthDateOnFile
	claim.Status = GuardianClaimStatusPending
	claim.CreateTime = time.Now()

	return claim
}
