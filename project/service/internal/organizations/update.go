package organizations

import (
	"context"
	"errors"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
)

// ErrChildrenLimitBelowCurrentCount means the requested ChildrenLimit is
// lower than how many children the organization already has -- shrinking
// the plan would immediately put them over their own new limit, which
// would then block adding any further child (see
// project/service/internal/children.ValidateCreateMutation) with a
// confusing "limit reached" message despite them just having lowered it
// themselves.
var ErrChildrenLimitBelowCurrentCount = errors.New("children limit cannot be set below the organization's current number of children")

func Update(ctx context.Context, argument organizationspg.UpdateArgument) (res organizationspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "organizations.Update", argument)
	defer logger.Exit(ctx, "organizations.Update", &res, &err)

	var (
		usesMutation = &organizationspg.UpdateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if rollbackErr := RollbackUpdateMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	if err = FinalizeUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return organizationspg.UpdateResponse{
		Organization: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *organizationspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "organizations.InitiateUpdateMutation", &err)

	result, err := Select(ctx, organizationspg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Organizations) == 0 {
		return nil
	}

	current := result.Organizations[0]

	if mutation.Foundation.ChildrenLimit != nil && *mutation.Foundation.ChildrenLimit < current.ChildrenLimit {
		count, countErr := children.CountForOrganization(ctx, current.Id)
		if countErr != nil {
			return countErr
		}
		if *mutation.Foundation.ChildrenLimit < count {
			return ErrChildrenLimitBelowCurrentCount
		}
	}

	updateOrganization := mutation.Foundation.Populate(current)

	mutation.Transient = organizationspg.Transient{
		Current: result.Organizations[0],
		Future:  updateOrganization,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *organizationspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "organizations.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *organizationspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "organizations.RollbackUpdateMutation", &err)

	return FinalizeUpdateMutation(ctx, &organizationspg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: organizationspg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
