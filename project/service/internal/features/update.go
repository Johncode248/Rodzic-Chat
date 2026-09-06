package features

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
)

func Update(ctx context.Context, argument featurespg.UpdateArgument) (res featurespg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "features.Update", argument)
	defer logger.Exit(ctx, "features.Update", &res, &err)

	var (
		usesMutation = &featurespg.UpdateMutation{
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

	return featurespg.UpdateResponse{
		OrganizationFeature: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *featurespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "features.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "features.InitiateUpdateMutation", &err)

	result, err := Select(ctx, featurespg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.OrganizationFeatures) == 0 {
		return nil
	}

	updateFeature := mutation.Foundation.Populate(result.OrganizationFeatures[0])

	mutation.Transient = featurespg.Transient{
		Current: result.OrganizationFeatures[0],
		Future:  updateFeature,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *featurespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "features.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "features.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *featurespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "features.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "features.RollbackUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return FinalizeUpdateMutation(ctx, &featurespg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: featurespg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
