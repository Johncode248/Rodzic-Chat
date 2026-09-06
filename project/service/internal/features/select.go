package features

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
)

func Select(ctx context.Context, argument featurespg.SelectArgument) (res featurespg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "features.Select", argument)
	defer logger.Exit(ctx, "features.Select", &res, &err)

	var (
		usesMutation = &featurespg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return featurespg.SelectResponse{
		OrganizationFeatures: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *featurespg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "features.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "features.ProcessSelectMutation", &err)

	result, err := gormdb.Select[featurespg.OrganizationFeature](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
