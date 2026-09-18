package guardianclaims

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

var ErrClaimNotFound = errors.New("guardian claim not found")
var ErrClaimAlreadyResolved = errors.New("this guardian claim has already been resolved")

// Resolve is the only way a GuardianClaim's Status ever changes -- approve
// creates the ParentChild link (with the pickup rights that come with it)
// that PublicRegister deliberately withheld when the claim was created;
// reject just marks the claim, no link is ever created.
func Resolve(ctx context.Context, organizationId, claimId, resolvedBy string, approve bool) (err error) {
	result, err := Select(ctx, guardianclaimspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: claimId},
					{Key: "organization_id", Operator: "=", Value: organizationId},
				}},
			},
		},
	})
	if err != nil {
		return err
	}
	if len(result.GuardianClaims) == 0 {
		return ErrClaimNotFound
	}

	claim := result.GuardianClaims[0]
	if claim.Status != guardianclaimspg.GuardianClaimStatusPending {
		return ErrClaimAlreadyResolved
	}

	status := guardianclaimspg.GuardianClaimStatusRejected
	if approve {
		status = guardianclaimspg.GuardianClaimStatusApproved
	}

	if _, err = Update(ctx, guardianclaimspg.UpdateArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: claim.Id}}},
			},
		},
		Status:     &status,
		ResolvedBy: &resolvedBy,
	}); err != nil {
		return err
	}

	if !approve {
		return nil
	}

	_, err = parentchild.Create(ctx, parentchildpg.CreateArgument{
		ParentId: claim.ClaimantId,
		ChildId:  claim.ChildId,
		Relation: parentchildpg.RelationTypeGuardian,
	})
	return err
}
