package guardianclaimspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

// UpdateArgument is used internally only, to resolve a claim (approve or
// reject) -- there's no field-by-field editing of a claim's content.
type UpdateArgument struct {
	Query      databasepg.Query     `json:"query,omitempty" validate:"required"`
	Status     *GuardianClaimStatus `json:"status,omitempty"`
	ResolvedBy *string              `json:"resolved_by,omitempty"`
}

func (a UpdateArgument) Populate(claim GuardianClaim) GuardianClaim {

	if a.Status != nil {
		claim.Status = *a.Status
		now := time.Now()
		claim.ResolvedAt = &now
	}

	if a.ResolvedBy != nil {
		claim.ResolvedBy = *a.ResolvedBy
	}

	return claim
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
