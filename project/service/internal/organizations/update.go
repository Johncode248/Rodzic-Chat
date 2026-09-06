package organizations

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

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
			if err = RollbackUpdateMutation(ctx, usesMutation); err != nil {
				panic(err)
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

	updateOrganization := mutation.Foundation.Populate(result.Organizations[0])

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
